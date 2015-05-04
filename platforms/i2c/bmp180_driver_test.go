package i2c

import (
	"testing"
	//"time"

	"github.com/hybridgroup/gobot"
)

// --------- HELPERS
func initTestBMP180Driver() (driver *BMP180Driver) {
	driver, _ = initTestBMP180DriverWithStubbedAdaptor()
	return
}

func initTestBMP180DriverWithStubbedAdaptor() (*BMP180Driver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor("adaptor")
	return NewBMP180Driver(adaptor, "bot"), adaptor
}

// --------- TESTS

func TestNewBMP180Driver(t *testing.T) {
	// Does it return a pointer to an instance of BMP180Driver?
	var bm interface{} = NewBMP180Driver(newI2cTestAdaptor("adaptor"), "bot")
	_, ok := bm.(*BMP180Driver)
	if !ok {
		t.Errorf("NewBMP180Driver() should have returned a *BMP180Driver")
	}
}

// Methods
func TestBMP180DriverStart(t *testing.T) {
	bmp, adaptor := initTestBMP180DriverWithStubbedAdaptor()

	adaptor.i2cReadImpl = func() ([]byte, error) {
		return []byte{0x10, 0x11, 0x12, 0x13,
			0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x20, 0x21, 0x22, 0x23,
			0x24, 0x25, 0x26, 0x27, 0x28, 0x29,
			0x30, 0x31}, nil
	}
	gobot.Assert(t, len(bmp.Start()), 0)
	gobot.Assert(t, bmp.Calibration.ac1, uint16(4113))
	gobot.Assert(t, bmp.Calibration.ac2, uint16(4627))
}

func TestBMP180DriverTempPressure(t *testing.T) {
	bmp, adaptor := initTestBMP180DriverWithStubbedAdaptor()
	adaptor.i2cReadImpl = func() ([]byte, error) {
		return []byte{0x10, 0x11, 0x12, 0x13,
			0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x20, 0x21, 0x22, 0x23,
			0x24, 0x25, 0x26, 0x27, 0x28, 0x29,
			0x30, 0x31}, nil
	}
	gobot.Assert(t, len(bmp.Start()), 0)
	//<-time.After(10 * time.Millisecond)
	adaptor.i2cReadImpl = func() ([]byte, error) {
		return []byte{0x10, 0x10}, nil
	}
	bmp.ReadRawTemp()
	bmp.CalculateTemperature()
	//<-time.After(10 * time.Millisecond)
	gobot.Assert(t, bmp.Temperature, float32(116.58878))

	//<-time.After(10 * time.Millisecond)
	adaptor.i2cReadImpl = func() ([]byte, error) {
		return []byte{0x20, 0x20, 0x20}, nil
	}
	bmp.ReadRawPressure(0)
	bmp.CalculatePressure()
	//<-time.After(10 * time.Millisecond)
	gobot.Assert(t, bmp.Pressure, float32(50.007942))
}

func TestBMP180DriverHalt(t *testing.T) {
	mpu := initTestBMP180Driver()

	gobot.Assert(t, len(mpu.Halt()), 0)
}
