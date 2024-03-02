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

	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_HUMAN = 15398392
	COLOR_BATTLE_STAGE_PET   = 8599608

	COLOR_BATTLE_BLOOD_UPPER      = 9211135
	COLOR_BATTLE_BLOOD_LOWER      = 255
	COLOR_BATTLE_MANA_UPPER       = 16758653
	COLOR_BATTLE_MANA_LOWER       = 16740864
	COLOR_BATTLE_BLOOD_MANA_EMPTY = 65536

	COLOR_BATTLE_RECALL_BUTTON = 7694643
	COLOR_BATTLE_NAME_1        = 37083
	COLOR_BATTLE_NAME_2        = 37086
	COLOR_BATTLE_NAME_3        = 37087
	COLOR_BATTLE_NAME_4        = 37050
	COLOR_BATTLE_NAME_5        = 37008

	COLOR_WINDOW_SKILL_UNSELECTED   = 4411988
	COLOR_WINDOW_SKILL_BOTTOM_SPACE = 11575428

	COLOR_NS_INVENTORY_SLOT_EMPTY = 15793151
	COLOR_PR_INVENTORY_SLOT_EMPTY = 15202301

	COLOR_NS_INVENTORY_PIVOT = 15967
	COLOR_BS_INVENTORY_PIVOT = 15967
	COLOR_PR_INVENTORY_PIVOT = 11113016

	COLOR_PR_PRODUCE_BUTTON = 7683891
	COLOR_PR_NOT_PRODUCING  = 11390937

	COLOR_ITEM_CAN_NOT_BE_USED = 255
	COLOR_ITEM_BOMB_7B         = 10936306
	COLOR_ITEM_BOMB_8B         = 14614527 // 8388607, 4194303
	COLOR_ITEM_BOMB_9A         = 30719    // 5626258
	COLOR_ITEM_POTION          = 16448250 // 8948665
)
