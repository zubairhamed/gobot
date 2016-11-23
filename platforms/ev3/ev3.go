package ev3


type ColorSensorMode string
type Ev3Port string
const (
	ColModeReflect = ColorSensorMode("COL-REFLECT")
	ColModeRgbRaw = ColorSensorMode("RGB-RAW")

	Port1 = Ev3Port("in1")
	Port2 = Ev3Port("in2")
	Port3 = Ev3Port("in3")
	Port4 = Ev3Port("in4")

	PortA = Ev3Port("outA")
	PortB = Ev3Port("outB")
	PortC = Ev3Port("outC")
	PortD = Ev3Port("outD")
)


