package items

import (
	"cg/game/enums"

	"github.com/g70245/win"
)

type Item struct {
	Name  string
	Color win.COLORREF
}

var (
	I_B_7B = Item{"7B", COLOR_ITEM_BOMB_7B}
	I_B_8B = Item{"8B", COLOR_ITEM_BOMB_8B}
	I_B_9A = Item{"9A", COLOR_ITEM_BOMB_9A}
	Bombs  = enums.GenericEnum[Item]{List: []Item{I_B_7B, I_B_8B, I_B_9A}}
)
