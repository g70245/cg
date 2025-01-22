package game

import (
	"cg/internal"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	PH_TELEPORTING        = []string{"被不可思", "你感覺到一股"}
	PH_OUT_OF_RESOURCE    = []string{"道具已經用完了"}
	PH_VERIFICATION       = []string{"驗證系統"}
	PH_ACTIVITY           = []string{"發現野生一級", "南瓜之王", "虎王", "釣魚途中"}
	PH_PRODUCTION_FAILURE = []string{}
)

const (
	DURATION_LOG_ACTIVITY        = 16 * time.Second
	DURATION_LOG_TELEPORTING     = 30 * time.Second
	DURATION_LOG_OUT_OF_RESOURCE = 30 * time.Second
	DURATION_LOG_VERIFICATION    = 5 * time.Second
)

func DoesEncounterActivityMonsters(gameDir string) bool {
	if gameDir == "" {
		return false
	}

	return doesPhraseExist(gameDir, 5, DURATION_LOG_ACTIVITY, PH_ACTIVITY)
}

func IsTeleported(gameDir string) bool {
	if gameDir == "" {
		return false
	}
	return doesPhraseExist(gameDir, 5, DURATION_LOG_TELEPORTING, PH_TELEPORTING)
}

func IsOutOfResource(gameDir string) bool {
	if gameDir == "" {
		return false
	}
	return doesPhraseExist(gameDir, 5, DURATION_LOG_OUT_OF_RESOURCE, PH_OUT_OF_RESOURCE)
}

func IsVerificationTriggered(gameDir string) bool {
	if gameDir == "" {
		return false
	}
	return doesPhraseExist(gameDir, 5, DURATION_LOG_VERIFICATION, PH_VERIFICATION)
}

func IsProductionStatusOK(name, gameDir string, before time.Duration) bool {
	if gameDir == "" {
		return false
	}
	return doesPhraseExist(gameDir, 10, before, append(PH_PRODUCTION_FAILURE, name))
}

func doesPhraseExist(gameDir string, lineCount int, before time.Duration, phrases []string) bool {
	logDir := fmt.Sprintf("%s/Log", gameDir)
	lines := internal.GetLastLines(logDir, lineCount)
	now := time.Now()
	for i := range lines {
		if len(lines[i]) < 9 {
			continue
		}

		h, hErr := strconv.Atoi(lines[i][1:3])
		m, mErrr := strconv.Atoi(lines[i][4:6])
		s, sErr := strconv.Atoi(lines[i][7:9])
		if hErr != nil || mErrr != nil || sErr != nil {
			continue
		}

		logTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.Local)
		for j := range phrases {
			if !logTime.Before(now.Add(-before)) && strings.Contains(lines[i], phrases[j]) {
				return true
			}
		}
	}
	return false
}
