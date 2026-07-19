package container

import (
	"errors"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/dialog"
)

const (
	alertMusicSelectionError      = "Could not select an alert music file."
	alertMusicInitializationError = "Could not initialize alert music. Make sure the selected file is a valid MP3 and that an audio output device is available."
	gameFolderSelectionError      = "Could not select the game folder."
	actionConfigSelectionError    = "Could not select an action configuration file."
	actionConfigLoadError         = "Could not load the selected action configuration. Make sure it is a valid .ac file."
	actionConfigDestinationError  = "Could not choose where to save the action configuration."
	actionConfigSaveError         = "Could not save the action configuration file."
	alertMusicSetupReminder       = "Select alert music."
	logAccessSetupReminder        = "Select a game folder."
	alertMusicAndLogSetupReminder = "Select alert music.\nSelect a game folder."
	invalidActionIDError          = "Enter a valid action ID."
	noAvailableActionIDError      = "No valid action ID is available."
)

func showErrorMessage(message string) {
	dialog.NewError(errors.New(message), window).Show()
}

func setupReminderMessage(alertMusicMissing, logAccessMissing bool) string {
	switch {
	case alertMusicMissing && logAccessMissing:
		return alertMusicAndLogSetupReminder
	case alertMusicMissing:
		return alertMusicSetupReminder
	case logAccessMissing:
		return logAccessSetupReminder
	default:
		return ""
	}
}

func validateActionID(value string, maximum int) error {
	if maximum < 1 {
		return errors.New(noAvailableActionIDError)
	}

	actionID, err := strconv.Atoi(value)
	if err != nil {
		return errors.New(invalidActionIDError)
	}
	if actionID < 1 || actionID > maximum {
		return fmt.Errorf("Action ID must be between 1 and %d.", maximum)
	}
	return nil
}
