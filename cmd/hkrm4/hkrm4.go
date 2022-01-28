package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"

	"github.com/benpye/hkrm4/internal/broadlink"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type FanConfig struct {
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

type Config struct {
	IP   net.IP `json:"ip"`
	MAC  string `json:"mac"`
	Type int    `json:"type"`

	Fans []FanConfig `json:"fans"`
}

func main() {
	configPath := flag.String("config", "config.json", "Path of config file.")
	data := flag.String("data", "data", "Path to store persistent data.")
	port := flag.String("port", "", "Listening port - by default randomised.")
	verbose := flag.Bool("verbose", false, "Verbose logging.")

	flag.Parse()

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	configDecoder := json.NewDecoder(configFile)

	var config Config
	err = configDecoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	mac, err := net.ParseMAC(config.MAC)
	if err != nil {
		log.Fatal(err)
	}

	bl, err := broadlink.NewDevice(config.IP, mac, config.Type)
	if err != nil {
		log.Fatal(err)
	}

	newFan := func(config FanConfig) *accessory.Accessory {
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

		setSpeed := func(speed float64) {
			step := int(speed / stepVal)
			if *verbose {
				log.Printf("Setting speed to %f (%d)", speed, step)
			}

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
				setSpeed(speed.GetValue())
			} else {
				err = bl.SendData(config.Commands.Speed[0])
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

			err = bl.SendData(config.Commands.LightToggle)

			if err != nil {
				log.Printf("error: %v", err)
			}
		})

		acc.AddService(fan.Service)
		acc.AddService(light.Service)

		return acc
	}

	temp, hum, err := bl.CheckSensors()
	if err != nil {
		log.Fatal(err)
	}

	temperature := service.NewTemperatureSensor()
	temperature.CurrentTemperature.Float.SetValue(temp)
	temperature.CurrentTemperature.Float.OnValueRemoteGet(func() float64 {
		temp, _, err := bl.CheckSensors()
		if err != nil {
			log.Print(err)
		}

		return temp
	})

	humidity := service.NewHumiditySensor()
	humidity.CurrentRelativeHumidity.Float.SetValue(hum)
	humidity.CurrentRelativeHumidity.Float.OnValueRemoteGet(func() float64 {
		_, hum, err := bl.CheckSensors()
		if err != nil {
			log.Print(err)
		}

		return hum
	})

	var fans []*accessory.Accessory
	for _, fanConfig := range config.Fans {
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
		Port:        *port,
		StoragePath: *data,
	}

	transport, err := hc.NewIPTransport(transportConfig, bridge.Accessory, fans...)
	if err != nil {
		log.Fatal(err)
	}

	transport.Start()
}
