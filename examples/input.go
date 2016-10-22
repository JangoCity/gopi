/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE.md
*/

// This example shows you the events which are generated by input devices
// You should call it with an argument -mouse or -touchscreen which then
// binds to the appropriate device
package main

import (
	"os"
	"fmt"
	"flag"
)

import (
	"../input"
	"../device/touchscreen/ft5406"
	"../device/mouse"
)

////////////////////////////////////////////////////////////////////////////////

var (
	flagTouchscreen = flag.Bool("touchscreen",false,"Bind input to touchscreen")
	flagMouse = flag.Bool("mouse",false,"Bind input to touchscreen")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	flag.Parse()

	if *flagTouchscreen == *flagMouse {
		fmt.Println("Invalid: use either -touchscreen or -mouse")
		os.Exit(-1)
	}

	var device *input.Device
	var err error
	if *flagTouchscreen {
		device, err = input.Open(ft5406.Config{})
	} else if *flagMouse {
		device, err = input.Open(mouse.Config{})
	}
	if err != nil {
		fmt.Println("Error: ",err)
		os.Exit(-1)
	}
	defer device.Close()

	fmt.Println("Device:",device)

	err = device.ProcessEvents(func(dev *input.Device, evt *input.InputEvent) {
		fmt.Println(evt)
	})
	if err != nil {
		fmt.Println("Error: ",err)
		os.Exit(-1)
	}
}
