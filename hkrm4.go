package main

import (
	"log"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/rob121/broadlinkgo"
)

const LightToggle = "b1c07c02ce9e06000c0d0d0d0d0d0e0d0d0d0d0d0d0d0d0d0d0e0d0d0d0d0daa1b0d0d1a1a0d1a0e1a0d0d1b0c1b0c1b1a0e0c1b190e1a0e190e190f0b1c0c1b190e0d1a1a0e0c1b1a0e190e0c1b1a0d0d1b190e1a0d1a0e0c1b0d1a0d1b1a0d0d1a1a0e190e190e1a0d1a0e190e1a0d0d1b0c1b0d1a0d1b1a0d0d1a0d1b0c1b0d1a0d1b0c1b0d1a0d1b190e1a0d1a0e0c1b1a0d1a0d1a0e1a0d1a0e190e190e1a0e0c0003490d0d0e0d0d0d0d0d0d0d0d0d0d0e0d0d0d0d0d0d0d0d0eaa1a0d0d1a1b0d1a0e190e0c1b0d1a0d1b1a0d0d1a1a0e190e1a0d1a0e0c1b0c1b1a0e0c1b1a0d0d1b190e1a0d0d1b190e0c1b1a0d1a0e190e0d1a0d1b0d1a1a0d0d1b190e1a0d1a0e1a0d1a0d1a0e190e0d1a0d1b0c1b0d1a1a0e0c1b0d1a0d1a0d1b0d1a0d1a0d1b0c1b1a0d1a0e1a0d0d1a1a0e190e1a0d1a0e190e1a0e190e190e0c0003a70e0c0e0c0e0d0d0d0d0d0d0e0d0d0d0d0d0d0d0d0d0e0daa1a0d0e1a1a0d1a0d1a0e0d1a0d1a0d1b1a0d0d1a1a0e1a0d1a0e190e0c1b0d1b190e0c1b1a0e0c1b1a0d1a0e0c1b190e0c1b1a0e190e1a0d0d1b0c1b0c1b1a0e0c1b1a0d1a0e190e1a0d1a0e190e190e0d1b0c1b0c1b0d1b0c1b0c1b0d1a1a0e0c1b0c1b1a0e190e0d1a1a0e190e1a0d1a0e190e190e0d1b190e190e0d1b0c1b190e0d0003480e0d0e0c0e0c0e0d0d0d0d0d0d0d0d0d0e0d0d0d0d0d0dab1a0d0d1a1a0e1a0d1a0d0d1b0c1b0d1a1a0e0d1a1a0d1a0e190e1a0e0c1b0c1b1a0d0d1b190e0c1b1a0e190e0c1b1a0e0c1b1a0d1a0e190e0c1b0d1b0c1b190e0d1b190e190e1a0d1a0e190e1a0d1a0e0c1b0d1a0d1b0c1b0c1b0d1b0c1b190e0d1b0c1b1a0d1a0d0d1b190e1a0d1a0e190e1a0d1a0e0c1b1a0d1a0e0c1b0d1a1a0e0c0005dc00000000000000000000"
const Speed0 = "b1c07c02ce9e06000d0c0e0d0d0d0e0c0e0c0e0d0d0d0d0d0e0c0e0d0d0d0dab1a0d0d1a1a0e190e1a0d0d1b0c1b0c1b1a0d0d1b190e1a0d1a0e190e0d1a0d1b190e0d1a1a0e0c1b1a0d1a0e0c1b1a0d0d1b190e1a0d1a0e0c1b0c1b0d1a1a0e0c1b1a0d1a0e1a0d1a0d1a0e190e1a0d0d1b0c1b0d1a0d1b0c1b0d1a0d1a0d1b0d1a0d1a0d1b0d1a1a0d1a0e1a0d1a0e190e190e1a0e190e190e1a0e190e190e0d1b0c0003490e0c0e0c0e0c0e0d0d0d0d0d0d0d0e0d0d0d0d0d0d0d0dab1a0d0d1a1b0d1a0d1a0e0c1b0d1a0d1b1a0d0d1a1a0d1a0e1a0d1a0e0c1b0d1a1a0e0c1b1a0d0d1a1a0e1a0d0d1a1a0e0c1b1a0d1a0e190e0d1a0d1b0c1b1a0d0d1b190e1a0d1a0e190e1a0d1a0e190e0c1b0d1a0d1b0c1b0d1a0d1b0c1b0d1a0d1b0c1b0d1a0d1b190e1a0d1a0e190e1a0d1a0e190e1a0d1a0e190e1a0d1a0e0c1b0c0003850d0d0e0c0e0c0e0c0e0d0d0d0e0c0e0d0d0d0d0d0d0d0daa1b0d0d1a1b0d1a0d1a0e0c1b0d1a0d1a1a0e0d1a1a0d1a0e190e1a0d0d1b0c1b1a0d0d1b190e0d1a1a0e190e0d1a1a0e0c1b1a0d1a0d1a0e0c1b0d1a0d1b1a0d0d1a1a0e1a0d1a0d1a0e190e1a0d1a0e0c1b0d1a0d1b0c1b0d1a0d1a0d1b1a0d0d1a0d1b0d1a1a0e0c1b1a0d1a0d1a0e1a0d1a0d1a0e0c1b1a0e190e190e0d1a1a0e0c0003490e0c0e0d0d0d0e0c0e0c0e0c0e0d0d0d0d0d0d0e0d0d0daa1a0d0e1a1a0d1a0e1a0d0d1a0d1b0c1b1a0d0d1b190e1a0d1a0e190e0c1b0d1b190e0c1b1a0d0d1b1a0d1a0d0d1b1a0d0d1a1a0e1a0d1a0d0d1b0c1b0d1a1a0e0c1b1a0d1a0e190e1a0d1a0e190e190e0d1b0c1b0c1b0d1a0d1b0c1b0d1a1a0e0c1b0d1a0d1b190e0d1a1a0d1a0e1a0d1a0e190e1a0d0d1b190e1a0d1a0e0c1b1a0d0d0005dc00000000000000000000"
const Speed1 = "b1c04201ce9e06000f0404070e0c0e0c0e0c0e0d0d0d0d0d0d0d0e0d0d0d0d0d0dab1a0d0d1a1a0e1a0d1a0e0c1b0c1b0d1a1a0e0c1b1a0d1a0e190e1a0d0d1b0c1b1a0d0d1b190e0d1a1a0e190e0c1b1a0e0c1b190e1a0e190e0c1b0d1a0d1b190e0d1a1a0e190e1a0d1a0e190e190e1a0e0c1b0d1a0d1b0c1b0d1a0d1a0d1b0d1a0d1a0d1b0d1a1a0d0d1b1a0d1a0d1a0e1a0d1a0d1a0e190e1a0d1a0e190e0c1b1a0e0c0003490e0c0e0d0d0d0d0d0e0c0e0c0e0d0d0d0d0d0d0d0d0e0daa1a0d0e1a1a0d1a0d1b0d0d1a0d1b0c1b1a0d0d1a1a0e190e1a0e190e0c1b0d1b190e0c1b1a0d0d1b190e1a0d0d1b190e0d1a1a0e190e1a0d0d1b0c1b0d1a1a0e0c1b190e1a0e190e190e1a0e190e190e0d1a0d1b0c1b0d1a0d1b0c1b0d1a0d1b0c1b0d1a0d1b1a0d0d1a1a0d1a0e1a0d1a0e190e1a0d1a0e190e190e1a0e0c1b1a0d0d0005dc00000000"
const Speed2 = "b1c03a01ce9e06006d090e0c0e0d0d0d0d0d0e0c0e0d0d0d0daa1b0d0d1a1a0e190e190e0d1b0c1b0c1b1a0e0c1b190e1a0d1a0e190e0d1a0d1b190e0d1a1a0e0c1b1a0d1a0e0c1b1a0d0d1b190e1a0d1a0e0c1b0c1b0d1a1a0e0d1a1a0d1a0e190e1a0d1a0e190e1a0e0c1b0c1b0d1a0d1b0c1b0d1a0d1b1a0d0d1a0d1b0c1b0d1a0d1a1a0e1a0d1a0d1a0e1a0d1a0e0c1b1a0d1a0e190e190e1a0e0c0003490d0d0e0c0e0d0d0d0d0d0d0d0d0e0d0d0d0d0d0d0e0c0eaa1a0d0d1a1b0d1a0d1a0e0c1b0d1a0d1b1a0d0d1a1a0e1a0d1a0d1a0e0c1b0d1a1a0e0c1b1a0d0d1a1a0e1a0d0d1a1a0e0c1b1a0d1a0e190e0d1a0d1b0c1b1a0d0d1b1a0d1a0d1a0e190e1a0d1a0e190e0c1b0d1a0d1b0c1b0d1a0d1b0c1b1a0d0d1b0d1a0d1a0d1b0c1b1a0d1a0e190e1a0d1a0e190e0c1b1a0e190e1a0d1a0e190e0c0005dc000000000000000000000000"
const Speed3 = "b1c04001ce9e06000c0d0d0d0d0d0d0d0e0d0d0d0d0d0d0d0d0d0d0e0d0d0daa1a0e0d1a1a0d1a0e190e0d1a0d1b0c1b1a0d0d1b190e1a0e190e190e0c1c0c1b190e0c1b1a0e0c1b1a0d1a0e0c1b1a0d0d1b190e1a0d1a0e0c1b0c1b0d1a1a0e0c1b1a0d1a0e190e1a0d1a0e190e1a0d0d1b0c1b0d1a0d1b0c1b0d1a1a0d0d1b0d1a0d1a0d1b0d1a0d1b190e1a0d1a0d1a0e1a0d0d1a1a0e1a0d1a0d1a0e190e190f0c0003490d0d0e0c0e0c0e0d0d0d0d0d0e0c0e0d0d0d0d0d0d0d0dab1a0d0d1a1b0d1a0d1a0e0d1a0d1a0d1a1a0e0c1b1a0e190e1a0d1a0e0c1b0c1b1a0e0c1b190e0c1c190e190e0d1b190e0c1b1a0d1a0e190e0c1b0d1b0c1b1a0d0d1b190e1a0d1a0e190e190e1a0e190e0c1b0d1a0d1b0c1b0d1a0d1b1a0d0d1a0d1b0c1b0d1a0d1b0c1b1a0d1a0e190e1a0d1a0e0c1b1a0d1a0e190e1a0d1a0d1a0e0c0005dc000000000000"

func main() {
	bl := broadlinkgo.NewBroadlink()
	err := bl.AddManualDevice("192.168.10.78", "ec:0b:ae:23:f2:78", 0x649b)
	if err != nil {
		log.Fatal(err)
	}

	info := accessory.Info{Name: "Ceiling fan"}
	acc := accessory.New(info, accessory.TypeFan)

	fan := service.NewFan()
	fan.On.SetValue(false)

	speed := characteristic.NewRotationSpeed()
	// speed.SetMaxValue(100.0)
	// speed.SetMinValue(0.0)
	// speed.SetStepValue(100.0 / 3.0)

	speed.SetMaxValue(99.9)
	speed.SetMinValue(0.0)
	speed.SetStepValue(33.3)

	setSpeed := func(speed float64) {
		log.Printf("Setting speed to %f", speed)

		if speed > 99 {
			log.Print(3)
			err = bl.Execute("ec:0b:ae:23:f2:78", Speed3)
		} else if speed > 66 {
			log.Print(2)
			err = bl.Execute("ec:0b:ae:23:f2:78", Speed2)
		} else if speed > 33 {
			log.Print(1)
			err = bl.Execute("ec:0b:ae:23:f2:78", Speed1)
		} else {
			log.Print(0)
			err = bl.Execute("ec:0b:ae:23:f2:78", Speed0)
		}

		if err != nil {
			log.Printf("error: %v", err)
		}
	}

	speed.OnValueRemoteUpdate(setSpeed)
	fan.AddCharacteristic(speed.Characteristic)

	fan.On.OnValueRemoteUpdate(func(on bool) {
		log.Printf("fan on = %v", on)
		if on {
			setSpeed(speed.GetValue())
		} else {
			err = bl.Execute("ec:0b:ae:23:f2:78", Speed0)
		}

		if err != nil {
			log.Printf("error: %v", err)
		}
	})

	light := service.NewLightbulb()

	light.On.SetValue(false)
	light.On.OnValueRemoteUpdate(func(on bool) {
		log.Printf("light on = %v", on)
		err = bl.Execute("ec:0b:ae:23:f2:78", LightToggle)

		if err != nil {
			log.Printf("error: %v", err)
		}
	})

	acc.AddService(fan.Service)
	acc.AddService(light.Service)

	config := hc.Config{}
	transport, err := hc.NewIPTransport(config, acc)
	if err != nil {
		log.Fatal(err)
	}

	transport.Start()
}
