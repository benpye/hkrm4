package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/benpye/hkrm4/internal/broadlink"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type fanConfig struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Manufacturer     string `json:"manufacturer"`
	Model            string `json:"model"`
	FirmwareRevision string `json:"firmwareRevision"`
	SerialNumber     string `json:"serialNumber"`
	Commands         struct {
		LightToggle []byte   `json:"lightToggle"`
		Speed       [][]byte `json:"speed"`
	} `json:"commands"`
}

type config struct {
	IP   net.IP `json:"ip"`
	MAC  string `json:"mac"`
	Type int    `json:"type"`

	Fans []fanConfig `json:"fans"`
}

type sensorCollector struct {
	bl                *broadlink.Device
	humidityMetric    *prometheus.Desc
	temperatureMetric *prometheus.Desc
}

var verbose *bool

func newSensorCollector(bl *broadlink.Device) *sensorCollector {
	return &sensorCollector{
		bl:                bl,
		humidityMetric:    prometheus.NewDesc("sensor_relative_humidity_percent", "Relative humidity in percent.", nil, nil),
		temperatureMetric: prometheus.NewDesc("sensor_temperature_celsius", "Temperature in degrees celsius.", nil, nil),
	}
}

func (c *sensorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.humidityMetric
	ch <- c.temperatureMetric
}

//Collect implements required collect function for all promehteus collectors
func (c *sensorCollector) Collect(ch chan<- prometheus.Metric) {
	if *verbose {
		log.Print("collecting sensor metrics")
	}

	temp, hum, err := c.bl.CheckSensors()
	if err != nil {
		log.Print(err)
		return
	}

	if *verbose {
		log.Printf("temperature: %f, humidity: %f", temp, hum)
	}

	ch <- prometheus.MustNewConstMetric(c.humidityMetric, prometheus.GaugeValue, hum)
	ch <- prometheus.MustNewConstMetric(c.temperatureMetric, prometheus.GaugeValue, temp)
}

func main() {
	configPath := flag.String("config", "config.json", "Path of config file.")
	data := flag.String("data", "data", "Path to store persistent data.")
	port := flag.String("port", "", "Listening port - by default randomised.")
	pin := flag.String("pin", "00102003", "PIN used for HomeKit pairing.")
	metricsPort := flag.String("metrics", "", "Metrics listening port - disabled if not specified.")
	verbose = flag.Bool("verbose", false, "Verbose logging.")

	flag.Parse()

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	configDecoder := json.NewDecoder(configFile)

	var cfg config
	err = configDecoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	mac, err := net.ParseMAC(cfg.MAC)
	if err != nil {
		log.Fatal(err)
	}

	bl, err := broadlink.NewDevice(cfg.IP, mac, cfg.Type)
	if err != nil {
		log.Fatal(err)
	}

	newFan := func(config fanConfig) *accessory.Accessory {
		info := accessory.Info{
			Name:             config.Name,
			Manufacturer:     config.Manufacturer,
			Model:            config.Model,
			FirmwareRevision: config.FirmwareRevision,
			SerialNumber:     config.SerialNumber,
		}

		acc := accessory.New(info, accessory.TypeFan)

		fan := service.NewFan()
		fan.On.SetValue(false)

		speed := characteristic.NewRotationSpeed()

		minVal := 0.0
		maxVal := 100.0

		numSteps := len(config.Commands.Speed) - 1
		stepVal := maxVal / float64(numSteps)

		speed.SetMaxValue(maxVal)
		speed.SetMinValue(minVal)
		speed.SetStepValue(stepVal)

		speedMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "service",
			Subsystem:   "fan",
			Name:        "speed_fraction",
			Help:        "Current fan speed.",
			ConstLabels: prometheus.Labels{"id": config.ID},
		})

		lightMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "service",
			Subsystem:   "light",
			Name:        "brightness_fraction",
			Help:        "Current light brightness.",
			ConstLabels: prometheus.Labels{"id": config.ID},
		})

		setSpeed := func(speed float64) {
			step := int(speed / stepVal)

			if *verbose {
				log.Printf("setting speed to %f (%d)", speed, step)
			}

			speedMetric.Set(speed)

			bl.SendData(config.Commands.Speed[step])

			if err != nil {
				log.Printf("error: %v", err)
			}
		}

		speed.OnValueRemoteUpdate(setSpeed)
		fan.AddCharacteristic(speed.Characteristic)

		fan.On.OnValueRemoteUpdate(func(on bool) {
			if *verbose {
				log.Printf("fan on = %v", on)
			}

			if on {
				setSpeed(speed.GetValue() / 100.0)
			} else {
				setSpeed(0)
			}

			if err != nil {
				log.Printf("error: %v", err)
			}
		})

		light := service.NewLightbulb()

		light.On.SetValue(false)
		light.On.OnValueRemoteUpdate(func(on bool) {
			if *verbose {
				log.Printf("light on = %v", on)
			}

			if on {
				lightMetric.Set(1.0)
			} else {
				lightMetric.Set(0.0)
			}

			err = bl.SendData(config.Commands.LightToggle)

			if err != nil {
				log.Printf("error: %v", err)
			}
		})

		acc.AddService(fan.Service)
		acc.AddService(light.Service)

		if *metricsPort != "" {
			err = prometheus.Register(speedMetric)
			if err != nil {
				log.Fatal(err)
			}

			prometheus.Register(lightMetric)
			if err != nil {
				log.Fatal(err)
			}
		}

		return acc
	}

	temp, hum, err := bl.CheckSensors()
	if err != nil {
		log.Fatal(err)
	}

	temperature := service.NewTemperatureSensor()
	temperature.CurrentTemperature.Float.SetValue(temp)
	temperature.CurrentTemperature.Float.OnValueRemoteGet(func() float64 {
		if *verbose {
			log.Print("query temperature")
		}

		temp, _, err := bl.CheckSensors()
		if err != nil {
			log.Print(err)
		}

		if *verbose {
			log.Printf("temperature = %f", temp)
		}

		return temp
	})

	humidity := service.NewHumiditySensor()
	humidity.CurrentRelativeHumidity.Float.SetValue(hum)
	humidity.CurrentRelativeHumidity.Float.OnValueRemoteGet(func() float64 {
		if *verbose {
			log.Print("query humidity")
		}

		_, hum, err := bl.CheckSensors()
		if err != nil {
			log.Print(err)
		}

		if *verbose {
			log.Printf("humidity = %f", hum)
		}

		return hum
	})

	var fans []*accessory.Accessory
	for _, fanConfig := range cfg.Fans {
		fans = append(fans, newFan(fanConfig))
	}

	info := accessory.Info{
		Name:             "BroadLink RM4 Pro",
		Manufacturer:     "BroadLink",
		Model:            "RM4 Pro",
		FirmwareRevision: "N/A",
		SerialNumber:     "N/A",
	}
	bridge := accessory.NewBridge(info)

	bridge.AddService(temperature.Service)
	bridge.AddService(humidity.Service)

	transportConfig := hc.Config{
		Pin:         *pin,
		Port:        *port,
		StoragePath: *data,
	}

	transport, err := hc.NewIPTransport(transportConfig, bridge.Accessory, fans...)
	if err != nil {
		log.Fatal(err)
	}

	if *metricsPort != "" {
		mux := http.NewServeMux()
		metricsServer := &http.Server{
			Addr:    ":" + *metricsPort,
			Handler: mux,
		}

		prometheus.MustRegister(newSensorCollector(bl))

		mux.Handle("/metrics", promhttp.Handler())
		go metricsServer.ListenAndServe()
	}

	transport.Start()
}
