package broadlink

import (
  "crypto/aes"
  "crypto/cipher"
  "errors"
  "fmt"
  "math/rand"
  "net"
)

type DeviceID []byte
type HardwareAddress []byte

type Device interface{
  MAC() HardwareAddress
  DeviceID() DeviceID
  PacketCount() uint16
  Encrypt( ps []byte ) ( []byte, error )
  Decrypt( ps []byte ) ( []byte, error )
}

type BaseDevice struct {
  conn  *net.UDPConn
  count uint16

  raddr *net.UDPAddr
  mac  HardwareAddress

  key []byte
  iv  []byte
  id  DeviceID
}

func ( bd *BaseDevice ) MAC() HardwareAddress {
  return bd.mac
}

func ( bd *BaseDevice ) DeviceID() DeviceID {
  return bd.id
}

func ( bd *BaseDevice ) PacketCount() uint16 {
  bd.count = ( bd.count + 1 ) & 0xffff
  return bd.count
}

func ( bd *BaseDevice ) Encrypt( ps []byte ) ( []byte, error ) {
  block, err := aes.NewCipher( bd.key )
  if err != nil {
    return []byte{}, err
  }
  ciphertext := make( []byte, len( ps ) )
  mode := cipher.NewCBCEncrypter( block, bd.iv )
  mode.CryptBlocks( ciphertext, ps )
  return ciphertext, nil
}

func ( bd *BaseDevice ) Decrypt( ps []byte ) ( []byte, error ) {
  block, err := aes.NewCipher(bd.key)
  if err != nil {
    return []byte{}, err
  }
  deciphertext := make( []byte, len( ps ) )
  mode := cipher.NewCBCDecrypter( block, bd.iv )
  mode.CryptBlocks( deciphertext, ps )
  return deciphertext, nil
}

func newBaseDevice( raddr *net.UDPAddr, mac HardwareAddress ) ( *BaseDevice, error ) {
  udpconn, err := net.ListenUDP( "udp", nil )
  if err != nil {
    return nil, err
  }

  return &BaseDevice{
    raddr: raddr,
    mac:  mac,

    conn: udpconn,
    key: []byte{ 0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02 },
    iv:  []byte{ 0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58 },
    id:  []byte{ 0, 0, 0, 0 },

    count: uint16( rand.Float64() * float64( 0xffff ) ),
  }, nil
}

func (bd *BaseDevice) newDevice(devtype uint16) (dev Device) {
  fmt.Printf("devtype:%x host:%s mac:%x\n", devtype, bd.raddr.String(), bd.MAC)

  switch devtype {
  // RM Mini
  case 0x2737:
    dev = newRM(bd)
  default:
  }

  return
}

func (bd *BaseDevice) Auth() error {
  command := NewAuthCommand()

  response, err := bd.SendCommand( command )
  if err != nil {
    return err
  }

  fmt.Printf( "auth resp data:%v\n", response )

  bd.id = response[0x00:0x04]
  bd.key = response[0x04:0x14]

  return nil
}

func ( bd *BaseDevice ) EnterLearning() ( []byte, error ) {
  command := NewEnterLearningCommand()

  response, err := bd.SendCommand( command )

  return response, err
}

func ( bd *BaseDevice ) CheckData() ( []byte, error ) {
  command := NewCheckDataCommand()

  response, err := bd.SendCommand( command )
  if err != nil {
    return response, err
  }

  return response[0x04:], err
}

func ( bd *BaseDevice ) SendData( data []byte ) error {
  command := NewSendDataCommand( data )

  _, err := bd.SendCommand( command )

  return err
}

func ( bd *BaseDevice ) SendCommand( command Command ) ( resp []byte, err error ) {
  resp = []byte{}

  cp := NewCommandPacket( bd, command )

  cps, err := cp.Bytes()
  if err != nil {
    return
  }

  _, err = bd.conn.WriteToUDP( cps, bd.raddr )
  if err != nil {
    return
  }

  resp = make( []byte, 1024 ) // Is 1024 too small?

  size, _, err := bd.conn.ReadFrom( resp )
  if err != nil {
    return
  }

  ps := resp[0:size]
  errcode := ps[0x22]
  if errcode != 0 {
    err = fmt.Errorf( "Response error code was non-zero: 0x%X", errcode )
    return
  }

  payload := ps[0x38:]
  if len(payload)  == 0 {
    err = errors.New( "Response contained empty payload" )
    return
  }

  payload, err = bd.Decrypt( payload )
  if err != nil {
    return
  }

  return payload, nil
}
