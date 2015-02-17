package i2c

import (
	"bytes"
	"encoding/binary"
	"time"
	"math"

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
	ac1 uint16
	ac2 uint16
	ac3 uint16
	ac4 uint16
	ac5 uint16
	ac6 uint16
	b1 uint16
	b2 uint16
	mb uint16
	mc uint16
	md uint16
}

type BMP180Driver struct {
	name          string
	connection    I2c
	interval      time.Duration
	Calibration 	BPMCalibrationData
	RawPressure 	uint16
	RawTemperature uint16
	Pressure   		float64
	Temperature   uint16
	Altitude			uint16
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

func (h *BMP180Driver) readRawPressure(mode uint8) (err error, pressure uint16) {
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

  h.RawPressure = uint16(((msb << 16) + (lsb << 8) + xlsb) >> (8-mode))
  return nil, h.RawPressure
}

func (h *BMP180Driver) calculateTemperature() uint16 {
  x1 := ((h.RawTemperature - h.Calibration.ac6) * h.Calibration.ac5) >> 15;
  x2 := (h.Calibration.mc << 11) / (x1 + h.Calibration.md)
  b5 := x1 + x2
  h.Temperature = ((b5 + 8) >> 4) / 10.0

  return b5
}

func (h *BMP180Driver) calculatePressure(mode uint16, b5 uint16) {
	var p float64
	b6 := b5 - 4000
	x1 := (h.Calibration.b2 * (b6 * b6) >> 12) >> 11
	x2 := (h.Calibration.ac2 * b6) >> 11
	x3 := x1 + x2
	b3 := math.Ceil(float64((((h.Calibration.ac1 * 4 + x3) << mode) + 2) / 4))

	x1 = (h.Calibration.ac3 * b6) >> 13
	x2 = (h.Calibration.b1 * ((b6 * b6) >> 12)) >> 16
	x3 = ((x1 + x2) + 2) >> 2
	b4 := (h.Calibration.ac4 * (x3 + 32768)) >> 15
	b7 := (h.RawPressure - uint16(b3)) * (50000 >> mode)

	if (b7 < 0x80000000) {
	    p = math.Ceil((b7 * 2) / b4)
	} else {
	    p = math.Ceil((b7 / b4) * 2)
	}

	x1 = (p >> 8) * (p >> 8)
	x1 = (x1 * 3038) >> 16
	x2 = (-7357 * p) >> 16

	p = p + ((x1 + x2 + 3791) >> 4)

	h.Pressure = p
}

func (h *BMP180Driver) calculateAltitude() {
	
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
