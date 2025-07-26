package enemy

type Position string

const (
	T1 Position = "T1"
	T2 Position = "T2"
	T3 Position = "T3"
	T4 Position = "T4"
	T5 Position = "T5"
	B1 Position = "B1"
	B2 Position = "B2"
	B3 Position = "B3"
	B4 Position = "B4"
	B5 Position = "B5"
)

func (o Position) String() string {
	return string(o)
}
