package i2c

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/hybridgroup/gobot"
)

var _ gobot.Driver = (*BMP180Driver)(nil)

const BMP180_REGISTER_CALIBRATION = 0xAA
const BMP180_REGISTER_CONTROL = 0xF4
const BMP180_REGISTER_TEMPDATA = 0xF6
const BMP180_REGISTER_PRESSUREDATA = 0xF6
const BMP180_REGISTER_READTEMPCMD = 0x2E
const BMP180_REGISTER_READPRESSURECMD = 0x34

const BPM180_MODE_LOWRES = 0
const BPM180_MODE_MEDIUMRES = 1
const BPM180_MODE_HIGHRES = 2
const BPM180_MODE_UHIGHRES = 3

type BPMCalibrationData struct {
	ac1 int16
	ac2 int16
	ac3 int16
	ac4 uint16
	ac5 uint16
	ac6 uint16
	b1 int16
	b2 int16
	mb int16
	mc int16
	md int16
}

type BMP180Driver struct {
	name          string
	connection    I2c
	interval      time.Duration
	Calibration 	BPMCalibrationData
	RawPressure 	uint16
	Pressure   		int16
	RawTemperature uint16
	Temperature   int16
	gobot.Eventer
}

// NewBMP180Driver creates a new driver with specified name and i2c interface
func NewBMP180Driver(a I2c, name string, v ...time.Duration) *BMP180Driver {
	m := &BMP180Driver{
		name:       name,
		connection: a,
		interval:   10 * time.Millisecond,
		Eventer:    gobot.NewEventer(),
	}

	if len(v) > 0 {
		m.interval = v[0]
	}

	m.AddEvent(Error)
	return m
}

func (h *BMP180Driver) Name() string                 { return h.name }
func (h *BMP180Driver) Connection() gobot.Connection { return h.connection.(gobot.Connection) }

// Start writes initialization bytes and reads from adaptor
// using specified interval to pressure and temperature data
func (h *BMP180Driver) Start() (errs []error) {
	if err := h.initialize(); err != nil {
		return []error{err}
	}

	// gobot.Every(h.interval, func() {
	// 	if err := h.connection.I2cWrite([]byte{MPU6050_RA_ACCEL_XOUT_H}); err != nil {
	// 		gobot.Publish(h.Event(Error), err)
	// 		return
	// 	}

	// 	ret, err := h.connection.I2cRead(14)
	// 	if err != nil {
	// 		gobot.Publish(h.Event(Error), err)
	// 		return
	// 	}
	// 	buf := bytes.NewBuffer(ret)
	// 	binary.Read(buf, binary.BigEndian, &h.Accelerometer)
	// 	binary.Read(buf, binary.BigEndian, &h.Gyroscope)
	// 	binary.Read(buf, binary.BigEndian, &h.Temperature)
	// })
	return
}

// Halt returns true if devices is halted successfully
func (h *BMP180Driver) Halt() (errs []error) { return }

func (h *BMP180Driver) initialize() (err error) {
	if err = h.connection.I2cStart(0x77); err != nil {
		return
	}

	if err = h.connection.I2cWrite([]byte{BMP180_REGISTER_CALIBRATION}); err != nil {
		return
	}
	ret, err := h.connection.I2cRead(22)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(ret)

	binary.Read(buf, binary.BigEndian, &h.Calibration.ac1)
	binary.Read(buf, binary.BigEndian, &h.Calibration.ac2)
	binary.Read(buf, binary.BigEndian, &h.Calibration.ac3)
	binary.Read(buf, binary.BigEndian, &h.Calibration.ac4)
	binary.Read(buf, binary.BigEndian, &h.Calibration.ac5)
	binary.Read(buf, binary.BigEndian, &h.Calibration.ac6)

	binary.Read(buf, binary.BigEndian, &h.Calibration.b1)
	binary.Read(buf, binary.BigEndian, &h.Calibration.b2)

	binary.Read(buf, binary.BigEndian, &h.Calibration.mb)
	binary.Read(buf, binary.BigEndian, &h.Calibration.mc)
	binary.Read(buf, binary.BigEndian, &h.Calibration.md)

	return nil
}

func (h *BMP180Driver) readRawTemp() (err error, temperature uint16) {
	if err := h.connection.I2cWrite([]byte{BMP180_REGISTER_CONTROL, BMP180_REGISTER_READTEMPCMD}); err != nil {
		gobot.Publish(h.Event(Error), err)
		return err, 0
	}
	<-time.After(5 * time.Millisecond)

	if err := h.connection.I2cWrite([]byte{BMP180_REGISTER_TEMPDATA}); err != nil {
		gobot.Publish(h.Event(Error), err)
		return err, 0
	}

	ret, err := h.connection.I2cRead(2)
	if err != nil {
		return err, 0
	}
	buf := bytes.NewBuffer(ret)

	binary.Read(buf, binary.BigEndian, &h.RawTemperature)
	return nil, h.RawTemperature
}

func (h *BMP180Driver) readRawPressure(mode uint8) (err error, pressure int16) {
	if err := h.connection.I2cWrite([]byte{BMP180_REGISTER_CONTROL, BMP180_REGISTER_READPRESSURECMD}); err != nil {
		gobot.Publish(h.Event(Error), err)
		return err, 0
	}
	<-time.After(waitTime(mode) * time.Millisecond)

	if err := h.connection.I2cWrite([]byte{BMP180_REGISTER_PRESSUREDATA}); err != nil {
		gobot.Publish(h.Event(Error), err)
		return err, 0
	}

	ret, err := h.connection.I2cRead(3)
	if err != nil {
		return err, 0
	}
	
	msb:= ret[0]
	lsb:= ret[1]
	xlsb := ret[2]

  var rawPress int16
  rawPress = int16(((msb << 16) + (lsb << 8) + xlsb) >> (8-mode))
  return nil, rawPress
}

func waitTime(mode uint8) time.Duration {
	switch mode {
	  case BPM180_MODE_LOWRES:
	    return 5
	  case BPM180_MODE_MEDIUMRES:
	    return 8
	  case BPM180_MODE_HIGHRES:
	    return 14
	  case BPM180_MODE_UHIGHRES:
	    return 26
	  default:
	    return 8
	}
}
