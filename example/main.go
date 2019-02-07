package main

import (
  "fmt"
  "time"
  // "encoding/base64"
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
    macString := dev.MAC().String()

    switch rmdev := dev.(type) {
    case *broadlink.RMProDevice:
      err = rmdev.BaseDevice.Auth()
      if err != nil {
        panic(err)
      }

      fmt.Printf( "RM Pro Device ID: %v (%v)\n", rmdev.DeviceID(), macString )
      time.Sleep( 500 * time.Millisecond )

      temp, err := rmdev.CheckTemperature()
      fmt.Printf( "Temperature: %f %v\n", temp, err )
      time.Sleep( 500 * time.Millisecond )

    case *broadlink.RMMiniDevice:
      err = rmdev.BaseDevice.Auth()
      if err != nil {
        panic(err)
      }
      
      fmt.Printf( "RM Mini Device ID: %v (%v)\n", rmdev.DeviceID(), macString )
    }

    time.Sleep( 200 * time.Millisecond )

    // if macString == "78:0f:77:00:ac:06" {
    //   rmdev := dev.(*broadlink.RMProDevice)

    //   _, err := rmdev.BaseDevice.EnterLearning()
    //   if err != nil {
    //     panic(err)
    //   }

    //   var data []byte
    //   for {
    //     data, err = rmdev.BaseDevice.CheckData()
    //     if err == nil {
    //       break
    //     }
    //     time.Sleep( 200 * time.Millisecond )
    //   }

    //   encoded := base64.StdEncoding.EncodeToString( data )
    //   fmt.Printf( "Learnt Data: %v\n", encoded )

    //   time.Sleep( 5000 * time.Millisecond )
    //   err = rmdev.BaseDevice.SendData( data )
    //   if err != nil {
    //     panic(err)
    //   }
    // }
  }
}
