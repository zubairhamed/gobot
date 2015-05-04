package main

import (
	"fmt"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/firmata"
	"github.com/hybridgroup/gobot/platforms/i2c"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("firmata", "/dev/ttyACM0")
	bpm180 := i2c.NewBMP180Driver(firmataAdaptor, "bpm180")

	work := func() {
		gobot.Every(1*time.Second, func() {
			fmt.Println("Pressure", bpm180.Pressure)
			fmt.Println("Temperature", bpm180.Temperature)
		})
	}

	robot := gobot.NewRobot("bpm180Bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{bpm180},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}