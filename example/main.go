package main

import (
	"github.com/woodnathan/broadlink"
	"time"
)

func main() {
	manager, err := broadlink.NewManager()
	if err != nil {
		panic(err)
	}

	devs, err := manager.Discover( 5 * time.Second )
	if err != nil {
		panic(err)
	}

	for _, dev := range devs {
		rmdev := dev.(*broadlink.RmDevice)
		err = rmdev.BaseDevice.Auth()
		if err != nil {
			panic(err)
		}
	}
}
