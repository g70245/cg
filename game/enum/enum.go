package enum

import (
	"fmt"
)

type GenericEnum[T any] struct {
	List []T
}

func (ge GenericEnum[T]) GetOptions() []string {
	var options []string
	for i := range ge.List {
		options = append(options, fmt.Sprint(ge.List[i]))
	}
	return options
}
