package broadlink

type Packet interface {
  Bytes() ( []byte, error )
}

func computeChecksum( ps []byte ) uint16 {
  checksum := uint16( 0xbeaf )
  for _, b := range ps {
    checksum += uint16( b )
  }
  checksum = checksum & 0xffff
  return checksum
}
