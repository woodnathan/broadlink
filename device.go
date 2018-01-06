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
  Encrypt( ps []byte ) ( []byte, error )
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
  payload := make([]byte, 0x50)
  payload[0x04] = 0x31
  payload[0x05] = 0x31
  payload[0x06] = 0x31
  payload[0x07] = 0x31
  payload[0x08] = 0x31
  payload[0x09] = 0x31
  payload[0x0a] = 0x31
  payload[0x0b] = 0x31
  payload[0x0c] = 0x31
  payload[0x0d] = 0x31
  payload[0x0e] = 0x31
  payload[0x0f] = 0x31
  payload[0x10] = 0x31
  payload[0x11] = 0x31
  payload[0x12] = 0x31
  payload[0x1e] = 0x01
  payload[0x2d] = 0x01
  payload[0x30] = 'T'
  payload[0x31] = 'e'
  payload[0x32] = 's'
  payload[0x33] = 't'
  payload[0x34] = ' '
  payload[0x35] = ' '
  payload[0x36] = '1'

  response, err := bd.SendPacket(0x65, payload)
  if err != nil {
    return err
  }

  fmt.Printf("auth resp:%v\n", response)

  //decode
  enc_payload := response[0x38:]
  if len(enc_payload) == 0 {
    return errors.New("auth failed!!!")
  }

  fmt.Printf("auth resp data:%v\n", response[:0x38])

  //aes = AES.new(bytes(self.key), AES.MODE_CBC, bytes(self.iv))
  //payload = aes.decrypt(bytes(enc_payload))
  block, aes_err := aes.NewCipher(bd.key)
  if aes_err != nil {
    return aes_err
  }
  mode := cipher.NewCBCDecrypter(block, bd.iv)
  respdata := make([]byte, len(enc_payload))
  mode.CryptBlocks(respdata, enc_payload)

  fmt.Printf("auth resp data:%v\n", respdata)

  bd.id = respdata[0x00:0x04]
  bd.key = respdata[0x04:0x14]

  return nil
}

func ( bd *BaseDevice ) SendPacket( command byte, payload []byte ) ( resp []byte, err error ) {
  bd.count = ( bd.count + 1 ) & 0xffff
  cp := NewCommandPacket( bd, uint16(command), bd.count, payload )

  cps, err := cp.Bytes()
  if err != nil {
    return []byte{}, err
  }

  _, werr := bd.conn.WriteToUDP( cps, bd.raddr )
  if werr != nil {
    err = werr
    return
  }

  resp = make( []byte, 1024 )

  size, raddr, rerr := bd.conn.ReadFrom( resp )
  if rerr != nil {
    err = rerr
    return
  }
  fmt.Printf("get %d bytes from %s\n", size, raddr.String())

  return resp[0:size], nil
}
