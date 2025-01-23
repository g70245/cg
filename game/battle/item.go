package battle

import (
	"cg/game"
	"cg/game/battle/enums"

	"github.com/g70245/win"
)

type Item struct {
	name  string
	color win.COLORREF
}

var (
	I_B_7B = Item{"7B", game.COLOR_ITEM_BOMB_7B}
	I_B_8B = Item{"8B", game.COLOR_ITEM_BOMB_8B}
	I_B_9A = Item{"9A", game.COLOR_ITEM_BOMB_9A}
	Bombs  = enums.GenericEnum[Item]{List: []Item{I_B_7B, I_B_8B, I_B_9A}}
)
