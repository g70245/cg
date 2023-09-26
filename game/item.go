package game

import (
	. "github.com/g70245/win"
)

type Item struct {
	name  string
	color COLORREF
}

type Items []Item

const (
	N_7B = "7B"
	N_8B = "8B"
	N_9A = "9A"
)

var (
	I_B_7B       = Item{N_7B, COLOR_ITEM_BOMB_7B}
	I_B_8B       = Item{N_8B, COLOR_ITEM_BOMB_8B}
	I_B_9A       = Item{N_9A, COLOR_ITEM_BOMB_9A}
	Bombs  Items = []Item{I_B_7B, I_B_8B, I_B_9A}
)

func (is Items) GetOptions() []string {
	options := make([]string, 0)
	for _, i := range is {
		options = append(options, i.name)
	}
	return options
}
