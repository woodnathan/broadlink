package broadlink

import (
  "net"
  "time"
)

type packetDiscovery struct {
  timestamp time.Time
  address *net.UDPAddr
}

type DiscoveryResponse struct {
  DeviceType uint16
  MAC []byte
}

func NewDiscoveryPacket( timestamp time.Time, address *net.UDPAddr ) Packet {
  return packetDiscovery{
    timestamp,
    address,
  }
}

func ( p packetDiscovery ) Bytes() ( []byte, error ) {
  ba := [0x30]byte{ 0 }

  // Separate out time components
  timestamp := p.timestamp
  _, timezone := timestamp.Zone()
  timezone = timezone / 3600 // seconds to hours
  year := timestamp.Year()

  // Set Timezone components
  if timezone < 0 {
    ba[0x08] = byte( 0xff + timezone - 1 )
    ba[0x09] = 0xff
    ba[0x0a] = 0xff
    ba[0x0b] = 0xff
  } else {
    ba[0x08] = byte( timezone )
    ba[0x09] = 0
    ba[0x0a] = 0
    ba[0x0b] = 0
  }

  // Set Time components
  ba[0x0c] = byte( year & 0xff )
  ba[0x0d] = byte( year >> 8 )
  ba[0x0e] = byte( timestamp.Minute() )
  ba[0x0f] = byte( timestamp.Hour() )
  ba[0x10] = byte( year % 1000 )
  ba[0x11] = byte( timestamp.Weekday() )
  ba[0x12] = byte( timestamp.Day() )
  ba[0x13] = byte( timestamp.Month() )

  // Set IP components
  ipv4 := p.address.IP.To4()
  if ipv4 == nil {
    panic( "IP address is not v4 address" ) // Should really complain earlier
  }
  ba[0x18] = byte( ipv4[0] )
  ba[0x19] = byte( ipv4[1] )
  ba[0x1a] = byte( ipv4[2] )
  ba[0x1b] = byte( ipv4[3] )
  ba[0x1c] = byte( p.address.Port & 0xff )
  ba[0x1d] = byte( p.address.Port >> 8 )
  ba[0x26] = 6

  checksum := computeChecksum( ba[:] )
  ba[0x20] = byte( checksum & 0xff )
  ba[0x21] = byte( checksum >> 8 )

  return ba[:], nil
}

func NewDiscoveryResponse( packet []byte ) DiscoveryResponse {
  devtype := uint16( packet[0x34] ) | uint16( packet[0x35] ) << 8
  mac := packet[0x3a:0x40]
  return DiscoveryResponse{
    DeviceType: devtype,
    MAC: mac,
  }
}
