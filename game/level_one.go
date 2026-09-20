package game

import (
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

func HasLevelOneEnemy(hWnd win.HWND) (bool, error) {
	return hasLevelOneEnemyWith(func(address uint32, size uint) ([]byte, error) {
		return internal.ReadMemoryAtAddress(hWnd, address, size)
	})
}

func hasLevelOneEnemyWith(readMemory processMemoryReader) (bool, error) {
	actors, err := readBattleEnemyActors(readMemory)
	if err != nil {
		return false, err
	}

	for index, actor := range actors {
		if actor == 0 {
			continue
		}

		level, err := readBattleUint32(readMemory, actor+MEMORY_ACTOR_LEVEL_OFFSET)
		if err != nil {
			slot := BATTLE_ENEMY_SLOT_START + index
			return false, fmt.Errorf("read battle slot %d actor level: %w", slot, err)
		}
		if level == 1 {
			return true, nil
		}
	}

	return false, nil
}
