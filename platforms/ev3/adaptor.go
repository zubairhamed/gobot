package ev3

func NewAdaptor(name string) *Ev3Adaptor {
	return &Ev3Adaptor{
		name: name,
	}
}

type Ev3Adaptor struct {
	name string
}

func (e *Ev3Adaptor) Name() string {
	return e.name
}

func (e *Ev3Adaptor) Connect() []error {
	return nil
}

func (e *Ev3Adaptor) Finalize() []error {
	return nil
}

func (e *Ev3Adaptor) Ping() string {
	return "pong"
}

