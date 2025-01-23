package controlunit

type ControlUnit string

const (
	Undefined ControlUnit = "Undefined"
	StartOver ControlUnit = "Start Over"
	Continue  ControlUnit = "Continue"
	Repeat    ControlUnit = "Repeat"
	Jump      ControlUnit = "Jump"
)

func (cu ControlUnit) String() string {
	return string(cu)
}
