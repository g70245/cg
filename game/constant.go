package game

const (
	GAME_WIDTH  = 640
	GAME_HEIGHT = 480

	ITEM_COL_LEN = 50
)

const (
	KEY_SKILL     uintptr = 0x57
	KEY_PET       uintptr = 0x52
	KEY_INVENTORY uintptr = 0x45
)

const (
	MEMORY_MAP_NAME  = 0x95C870
	MEMORY_MAP_POS_X = 0x95C88C
	MEMORY_MAP_POS_Y = 0x95C890
)

const (
	COLOR_ANY = 0

	COLOR_SCENE_NORMAL = 15595514
	COLOR_SCENE_BATTLE = 15595514

	COLOR_MENU_BUTTON_NORMAL     = 15135992
	COLOR_MENU_BUTTON_POPOUT     = 10331818
	COLOR_MENU_BUTTON_CONTACT    = 15201528
	COLOR_MENU_BUTTON_PET_POPOUT = 10331817
	COLOR_MENU_HIDDEN            = 7568253

	COLOR_WINDOW_SKILL_UNSELECTED   = 4411988
	COLOR_WINDOW_SKILL_BOTTOM_SPACE = 11575428

	COLOR_NS_INVENTORY_SLOT_EMPTY = 15793151

	COLOR_NS_INVENTORY_PIVOT = 15967
)
