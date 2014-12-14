package i2c

import (
	"testing"

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
	mpu := initTestBMP180Driver()

	gobot.Assert(t, len(mpu.Start()), 0)
}

func TestBMP180DriverHalt(t *testing.T) {
	mpu := initTestBMP180Driver()

	gobot.Assert(t, len(mpu.Halt()), 0)
}
