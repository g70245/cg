package container

import (
	"cg/game"
	"cg/utils"
	"time"

	"fyne.io/fyne/v2/dialog"
)

func notifyBeeperConfig(title string) {
	notifySetupConfig(title, !utils.Beeper.IsReady(), false)
}

func notifyLogConfig(title string) {
	notifySetupConfig(title, false, game.ValidateLogDirectory(r.getGameDir()) != nil)
}

func validateLogConfig(title string) bool {
	if err := game.ValidateLogDirectory(r.getGameDir()); err != nil {
		dialog.NewInformation(title, err.Error(), window).Show()
		return false
	}
	return true
}

func notifyBeeperAndLogConfig(title string) {
	notifySetupConfig(title, !utils.Beeper.IsReady(), game.ValidateLogDirectory(r.getGameDir()) != nil)
}

func notifySetupConfig(title string, alertMusicMissing, logAccessMissing bool) {
	notifySetupMessage(title, setupReminderMessage(alertMusicMissing, logAccessMissing))
}

func notifySetupMessage(title, message string) {
	if message == "" {
		return
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
		dialog.NewInformation(title, message, window).Show()
	}()
}
