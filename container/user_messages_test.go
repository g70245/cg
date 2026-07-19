package container

import "testing"

func TestSetupReminderMessage(t *testing.T) {
	tests := []struct {
		name              string
		alertMusicMissing bool
		logAccessMissing  bool
		want              string
	}{
		{name: "ready"},
		{name: "alert music", alertMusicMissing: true, want: "Select alert music."},
		{name: "log access", logAccessMissing: true, want: "Select a game folder."},
		{name: "both", alertMusicMissing: true, logAccessMissing: true, want: "Select alert music.\nSelect a game folder."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := setupReminderMessage(test.alertMusicMissing, test.logAccessMissing)
			if got != test.want {
				t.Fatalf("setupReminderMessage() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestValidateActionID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		maximum int
		wantErr string
	}{
		{name: "valid lower bound", value: "1", maximum: 3},
		{name: "valid upper bound", value: "3", maximum: 3},
		{name: "not a number", value: "abc", maximum: 3, wantErr: invalidActionIDError},
		{name: "below range", value: "0", maximum: 3, wantErr: "Action ID must be between 1 and 3."},
		{name: "above range", value: "4", maximum: 3, wantErr: "Action ID must be between 1 and 3."},
		{name: "no available ID", value: "1", maximum: 0, wantErr: noAvailableActionIDError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateActionID(test.value, test.maximum)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validateActionID() error = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != test.wantErr {
				t.Fatalf("validateActionID() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}
