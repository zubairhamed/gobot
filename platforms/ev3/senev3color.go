package ev3

import (
	"github.com/ev3go/ev3dev"
	"github.com/hybridgroup/gobot"
	"time"
)

func NewColorSensorDriver(adaptor *Ev3Adaptor, name string, port Ev3Port, mode ColorSensorMode) *ColorSensorDriver {
	return &ColorSensorDriver{
		name: name,
		connection: adaptor,
		interval: 500*time.Millisecond,
		halt: make(chan bool, 0),
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
		port: port,
		mode: mode,
	}
}

type ColorSensorDriver struct {
	name string
	connection gobot.Connection
	interval time.Duration
	halt chan bool
	gobot.Eventer
	gobot.Commander
	port Ev3Port
	sensor *ev3dev.Sensor
	mode ColorSensorMode
}

func (e *ColorSensorDriver) Read() []float64 {
	s := e.sensor
	n, err := s.NumValues()
	if err != nil {
		panic(err.Error())
	}

	fv := []float64{}

	for i := 0; i < n; i++ {
		v, err := s.Value(i)
		if err != nil {
			panic(err.Error())
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err.Error())
		}

		fv = append(fv, f)
	}
	return fv
}

func (e *ColorSensorDriver) Name() string {
	return e.name
}

func (e *ColorSensorDriver) Connection() gobot.Connection {
	return e.connection
}

func (e *ColorSensorDriver) adaptor() *Ev3Adaptor {
	return e.Connection().(*Ev3Adaptor)
}

func (e *ColorSensorDriver) Ping() string {
	return e.adaptor().Ping()
}

func (e *ColorSensorDriver) Start() []error {

	s, err := ev3dev.SensorFor(string(e.port), "lego-ev3-color")
	if err != nil {
		return []error { err }
	}

	s.SetMode(string(e.mode))

	e.sensor = s

	return nil
}

func (e *ColorSensorDriver) Halt() []error {
	e.halt <- true
	return nil
}

type ColorSensorValue struct {

}