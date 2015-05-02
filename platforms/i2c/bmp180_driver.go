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

type BPMPolynomials struct {
	c3 float64
	c4 float64
	b1 float64
	c5 float64
	c6 float64
	mc float64
	md float64
	x0 float64
	x1 float64
	x2 float64
	y0 float64
	y1 float64
	y2 float64
	p0 float64
	p1 float64
	p2 float64
}

type BMP180Driver struct {
	name          string
	connection    I2c
	interval      time.Duration
	Calibration 	BPMCalibrationData
	Polynomials		BPMPolynomials
	RawPressure 	uint16
	RawTemperature uint16
	Pressure   		float64
	Temperature   float64
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

	h.Polynomials.c3 = 160.0 * math.Pow(2,-15.0) * float64(h.Calibration.ac3)
	h.Polynomials.c4 = math.Pow(10, -3) * math.Pow(2,-15) * float64(h.Calibration.ac4)
	h.Polynomials.b1 = math.Pow(160, 2) * math.Pow(2, -30) * float64(h.Calibration.b1)
	h.Polynomials.c5 = (math.Pow(2, -15) / 160) * float64(h.Calibration.ac5)
	h.Polynomials.c6 = float64(h.Calibration.ac6)
	h.Polynomials.mc = (math.Pow(2, 11) / math.Pow(160, 2)) * float64(h.Calibration.mc)
	h.Polynomials.md = float64(h.Calibration.md) / 160.0
	h.Polynomials.x0 = float64(h.Calibration.ac1)
	h.Polynomials.x1 = 160.0 * math.Pow(2, -13) * float64(h.Calibration.ac2)
	h.Polynomials.x2 = math.Pow(160, 2) * math.Pow(2, -25) * float64(h.Calibration.b2)
	h.Polynomials.y0 = h.Polynomials.c4 * math.Pow(2, 15)
	h.Polynomials.y1 = h.Polynomials.c4 * h.Polynomials.c3
	h.Polynomials.y2 = h.Polynomials.c4 * h.Polynomials.b1
	h.Polynomials.p0 = (3791.0 - 8.0) / 1600.0
	h.Polynomials.p1 = 1.0 - 7357.0 * math.Pow(2, -20)
	h.Polynomials.p2 = 3038.0 * 100.0 * math.Pow(2, -36)

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

func (h *BMP180Driver) calculateTemperature() float64 {
	a := h.Polynomials.c5 * (float64(h.RawTemperature) - h.Polynomials.c6)
	h.Temperature = a + (h.Polynomials.mc / (a + h.Polynomials.md))
	return h.Temperature
}

func (h *BMP180Driver) calculatePressure() float64 {
	s := h.Temperature - 25.0;
	x := (h.Polynomials.x2 * math.Pow(s, 2)) + (h.Polynomials.x1 * s) + h.Polynomials.x0
	y := (h.Polynomials.y2 * math.Pow(s, 2)) + (h.Polynomials.y1 * s) + h.Polynomials.y0
	z := (float64(h.RawPressure) - x) / y
	h.Pressure = (h.Polynomials.p2 * math.Pow(z, 2)) + (h.Polynomials.p1 * z) + h.Polynomials.p0
	return h.Pressure
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
