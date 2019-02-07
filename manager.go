package broadlink

import (
  "log"
  "net"
  "time"
  "errors"
)

const DISCOVERY_PORT = 80

type Manager struct {
  conn *net.UDPConn
  laddr *net.UDPAddr
  devices []Device
}

func NewManager() ( *Manager, error ) {
  udpconn, err := net.ListenUDP( "udp", nil )
  if err != nil {
    return nil, err
  }

  laddr := GetLocalAddr()

  return &Manager{
    conn: udpconn,
    laddr: laddr,
  }, nil
}

func ( m *Manager ) Discover( timeout time.Duration ) ( devices []Device, err error ) {
  devices = make( []Device, 0 )

  mcaddr := &net.UDPAddr{
    IP:   net.IPv4bcast,
    Port: DISCOVERY_PORT,
  }

  // broadcast
  dp := NewDiscoveryPacket( time.Now(), m.laddr )
  dps, err := dp.Bytes()
  if err != nil {
    return devices, err
  }
  m.conn.WriteTo( dps, mcaddr )

  //read
  m.conn.SetReadDeadline( time.Now().Add( timeout ) )
  resp := make( []byte, 1024 )

  for {
    size, raddr, err := m.conn.ReadFromUDP( resp )
    if err, ok := err.(net.Error); ok && err.Timeout() {
      break
    }
    if err != nil {
      return devices, err
    }
    
    if size == 0 {
      err = errors.New( "Unable to read discovery response" )
      return devices, err
    }

    dr := NewDiscoveryResponse( resp )

    bd, err := newBaseDevice( raddr, dr.MAC )
    if err != nil {
      return devices, err
    }
    dev := bd.newDevice( dr.DeviceType )
    m.devices = append( m.devices, dev )

    //break // Use channels to push out new devices
  }

  return m.devices, nil
}

func GetLocalAddr() *net.UDPAddr {
  udpcon, err := net.DialUDP( "udp",
    nil,
    &net.UDPAddr{
      IP:   net.ParseIP("8.8.8.8"),
      Port: 53,
    },
  )
  if err != nil {
    log.Fatalf( "Failed to obtain local address: %s\n", err )
  }
  defer udpcon.Close()

  laddr := udpcon.LocalAddr()
  ludpaddr, err := net.ResolveUDPAddr( laddr.Network(), laddr.String() )
  if err != nil {
    log.Fatalf( "Failed to resolve local UDP address: %s\n", err )
  }
  return ludpaddr
}
