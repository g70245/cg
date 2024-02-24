package game

import (
	"github.com/g70245/win"
)

type Item struct {
	name  string
	color win.COLORREF
}

type Items []Item

var (
	I_B_7B       = Item{"7B", COLOR_ITEM_BOMB_7B}
	I_B_8B       = Item{"8B", COLOR_ITEM_BOMB_8B}
	I_B_9A       = Item{"9A", COLOR_ITEM_BOMB_9A}
	Bombs  Items = []Item{I_B_7B, I_B_8B, I_B_9A}
)

func (is Items) GetOptions() []string {
	options := make([]string, 0)
	for _, i := range is {
		options = append(options, i.name)
	}
	return options
}
