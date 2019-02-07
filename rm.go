package broadlink

// There's definitely a way to not implement both, but it's escaping me

type RMProDevice struct {
  *BaseDevice
}

func newRMPro(dev *BaseDevice) *RMProDevice {
  return &RMProDevice{
    BaseDevice: dev,
  }
}

func (rm *RMProDevice) Check() {

}

func (rm *RMProDevice) Send(data []byte) {

}

func (rm *RMProDevice) EnterLearning() {

}

func (rm *RMProDevice) CheckTemperature() ( float32, error ) {
  command := NewCheckDataCommand()

  response, err :=  rm.SendCommand( command )
  if err != nil {
    return 0.0, err
  }

  temp := ( float32( response[0x04] ) * 10 + float32( response[0x05] ) ) / 10.0

  return temp, nil
}

type RMMiniDevice struct {
  *BaseDevice
}

func newRMMini(dev *BaseDevice) *RMMiniDevice {
  return &RMMiniDevice{
    BaseDevice: dev,
  }
}

func (rm *RMMiniDevice) Check() {

}

func (rm *RMMiniDevice) Send(data []byte) {

}

func (rm *RMMiniDevice) EnterLearning() {

}
