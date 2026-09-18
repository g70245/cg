package container

import (
	"cg/game"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynecontainer "fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/g70245/win"
)

func TestRatioMonitorButtonText(t *testing.T) {
	if got := ratioMonitorButtonText("HP", 0.5); got != "HP: 50%" {
		t.Fatalf("ratioMonitorButtonText() = %q, want %q", got, "HP: 50%")
	}
	if got := ratioMonitorButtonText("Pet MP", 0.15); got != "Pet MP: 15%" {
		t.Fatalf("ratioMonitorButtonText() = %q, want %q", got, "Pet MP: 15%")
	}
}

func TestBattleGroupViewCompactModeKeepsSwitchAndRestoreButtons(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	switchButton := widget.NewButton("Start", nil)
	restoreButton := widget.NewButton("Restore", nil)
	fullMenuObjects := []fyne.CanvasObject{
		widget.NewButton("Monitoring", nil),
		widget.NewButton("Target Priority", nil),
		widget.NewButton("Load", nil),
		widget.NewButton("Delete", nil),
		switchButton,
	}
	menu := newBattleGroupMenu(fullMenuObjects, switchButton, restoreButton)
	navigation := newBattleNavigationView(game.Games{}, game.Games{}, func() string { return "" }, nil)
	testGames := game.Games{"1": win.HWND(1)}
	steps := new(atomic.Uint32)
	steps.Store(480)
	ridingSteps := newRidingStepsViewWith(testGames, testGames, func(win.HWND) (uint32, error) { return steps.Load(), nil }, time.Hour)
	view := newBattleGroupView(menu, navigation, ridingSteps, fynecontainer.NewVBox(widget.NewLabel("Worker settings")))
	defer view.close()

	view.setCompact(true)
	if got, want := len(view.container.Objects), len(view.compactObjects); got != want {
		t.Fatalf("compact group object count = %d, want %d", got, want)
	}
	if got := len(menu.container.Objects); got != 2 {
		t.Fatalf("compact menu object count = %d, want 2", got)
	}
	if menu.container.Objects[0] != switchButton {
		t.Fatal("compact menu replaced the existing start/stop button")
	}
	if menu.container.Objects[1] != restoreButton {
		t.Fatal("compact menu does not contain the restore button")
	}
	foundNavigation := false
	for _, object := range view.container.Objects {
		if object == navigation.container {
			foundNavigation = true
		}
	}
	if !foundNavigation {
		t.Fatal("compact group does not contain navigation")
	}
	foundRidingSteps := false
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			foundRidingSteps = true
		}
	}
	if !foundRidingSteps {
		t.Fatal("compact group does not contain riding steps")
	}
	steps.Store(0)
	ridingSteps.update()
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			t.Fatal("compact group retained an empty riding steps row")
		}
	}
	steps.Store(480)
	ridingSteps.update()
	foundRidingSteps = false
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			foundRidingSteps = true
		}
	}
	if !foundRidingSteps {
		t.Fatal("compact group did not restore riding steps row")
	}
	view.setSelected(false)
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			t.Fatal("unselected compact group retained its riding steps row")
		}
	}
	view.setSelected(true)
	foundRidingSteps = false
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			foundRidingSteps = true
		}
	}
	if !foundRidingSteps {
		t.Fatal("selected compact group did not restore its riding steps row")
	}

	view.setCompact(false)
	if got, want := len(view.container.Objects), len(view.fullObjects); got != want {
		t.Fatalf("full group object count = %d, want %d", got, want)
	}
	if !reflect.DeepEqual(menu.container.Objects, fullMenuObjects) {
		t.Fatal("full menu objects were not restored")
	}
	for _, object := range view.container.Objects {
		if object == navigation.container {
			t.Fatal("full group unexpectedly contains navigation")
		}
		if object == ridingSteps.container {
			t.Fatal("full group unexpectedly contains riding steps")
		}
	}
	steps.Store(480)
	ridingSteps.update()
	for _, object := range view.compactHeader.Objects {
		if object == ridingSteps.container {
			t.Fatal("riding steps update added its row to the full group header")
		}
	}
}
