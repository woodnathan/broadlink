package main

import (
  "fmt"
  "time"
  "github.com/woodnathan/broadlink"
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

  fmt.Printf( "Devices: %v\n", devs )

  for _, dev := range devs {
    rmdev := dev.(*broadlink.RmDevice)
    err = rmdev.BaseDevice.Auth()
    if err != nil {
      panic(err)
    }

    _, err := rmdev.BaseDevice.EnterLearning()
    if err != nil {
      panic(err)
    }

    var data []byte
    for {
      data, err = rmdev.BaseDevice.CheckData()
      if err == nil {
        break
      }
      time.Sleep( 200 * time.Millisecond )
    }
    fmt.Printf( "Learnt Data: %v\n", data )

    time.Sleep( 5000 * time.Millisecond )
    err = rmdev.BaseDevice.SendData( data )
    if err != nil {
      panic(err)
    }
  }
}
