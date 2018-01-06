package broadlink

type packetCommand struct {
  Device Device
  Command Command
  PacketCount uint16
}

func NewCommandPacket( device Device, command Command ) Packet {
  return packetCommand{
    Device: device,
    Command: command,
    PacketCount: device.PacketCount(),
  }
}

func ( p packetCommand ) Bytes() ( []byte, error ) {
  ba := [0x38]byte{ 0 }

  ba[0x00] = 0x5a
  ba[0x01] = 0xa5
  ba[0x02] = 0xaa
  ba[0x03] = 0x55
  ba[0x04] = 0x5a
  ba[0x05] = 0xa5
  ba[0x06] = 0xaa
  ba[0x07] = 0x55
  ba[0x24] = 0x2a
  ba[0x25] = 0x27
  ba[0x26] = byte( p.Command.Code() )
  ba[0x27] = 0 // Other half of command code?
  ba[0x28] = byte( p.PacketCount & 0xff )
  ba[0x29] = byte( p.PacketCount >> 8 )

  MAC := p.Device.MAC()
  ba[0x2a] = MAC[0]
  ba[0x2b] = MAC[1]
  ba[0x2c] = MAC[2]
  ba[0x2d] = MAC[3]
  ba[0x2e] = MAC[4]
  ba[0x2f] = MAC[5]

  deviceID := p.Device.DeviceID()
  ba[0x30] = deviceID[0]
  ba[0x31] = deviceID[1]
  ba[0x32] = deviceID[2]
  ba[0x33] = deviceID[3]

  payload, err := p.Command.Bytes()
  if err != nil {
    return []byte{}, err
  }
  
  if len( payload ) > 0 {
    numpad := 16 * ( ( len(payload) / 16 ) + 1 )
    deltapad := numpad - len( payload )

    for n := 0; n < deltapad; n++ {
      payload = append( payload, byte( 0 ) )
    }
  }

  payloadChecksum := computeChecksum( payload )
  ba[0x34] = byte( payloadChecksum & 0xff )
  ba[0x35] = byte( payloadChecksum >> 8 )

  encPayload, err := p.Device.Encrypt( payload )
  if err != nil {
    return []byte{}, err
  }

  bs := append( ba[:], encPayload... )

  packetChecksum := computeChecksum( bs )
  bs[0x20] = byte( packetChecksum & 0xff )
  bs[0x21] = byte( packetChecksum >> 8 )

  return bs, nil
}
