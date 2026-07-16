package container

import (
	"cg/game/battle"
	"cg/game/enum/controlunit"
	"cg/game/enum/role"
	"cg/game/enum/threshold"
	"fmt"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func generateTags(actionState battle.ActionState) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainers(actionState, role.Character)
	tagContaines = append(tagContaines, createTagContainers(actionState, role.Pet)...)
	return
}

var (
	characterFinishingTagColor   = color.RGBA{235, 206, 100, uint8(math.Round(1 * 255))}
	characterConditionalTagColor = color.RGBA{100, 206, 235, uint8(math.Round(1 * 255))}
	characterSpecialTagColor     = color.RGBA{206, 235, 100, uint8(math.Round(1 * 255))}
	petFinishingTagColor         = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	petConditionalTagColor       = color.RGBA{0, 79, 245, uint8(math.Round(0.8 * 255))}
	petSpecialTagColor           = color.RGBA{79, 245, 0, uint8(math.Round(0.8 * 255))}
)

func createTagContainers(actionState battle.ActionState, r role.Role) (tagContainers []fyne.CanvasObject) {
	var tag string
	var anyActions []any
	if r == role.Character {
		for _, hs := range actionState.GetCharacterActions() {
			anyActions = append(anyActions, hs)
		}
	} else {
		for _, ps := range actionState.GetPetActions() {
			anyActions = append(anyActions, ps)
		}
	}

	for _, action := range anyActions {
		var tagColor color.RGBA
		if r == role.Character {
			tag = action.(battle.CharacterAction).Action.String()
			switch {
			case strings.Contains(action.(battle.CharacterAction).Action.String(), "**"):
				tagColor = characterFinishingTagColor
			case strings.Contains(action.(battle.CharacterAction).Action.String(), "*"):
				tagColor = characterConditionalTagColor
			default:
				tagColor = characterSpecialTagColor
			}
		} else {
			tag = action.(battle.PetAction).Action.String()
			switch {
			case strings.Contains(action.(battle.PetAction).Action.String(), "**"):
				tagColor = petFinishingTagColor
			case strings.Contains(action.(battle.PetAction).Action.String(), "*"):
				tagColor = petConditionalTagColor
			default:
				tagColor = petSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)

		if r == role.Character {
			if offset := action.(battle.CharacterAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d:%d", tag, offset, action.(battle.CharacterAction).Level)
			}
		} else {
			if offset := action.(battle.PetAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d", tag, offset)
			}
		}

		var param string
		var threshold threshold.Threshold
		var successControlUnit controlunit.ControlUnit
		var successJumpId int
		var failureControlUnit controlunit.ControlUnit
		var failureJumpId int

		if r == role.Character {
			param = action.(battle.CharacterAction).Param
			threshold = action.(battle.CharacterAction).Threshold
			successControlUnit = action.(battle.CharacterAction).SuccessControlUnit
			failureControlUnit = action.(battle.CharacterAction).FailureControlUnit
			successJumpId = action.(battle.CharacterAction).SuccessJumpId
			failureJumpId = action.(battle.CharacterAction).FailureJumpId
		} else {
			param = action.(battle.PetAction).Param
			threshold = action.(battle.PetAction).Threshold
			successControlUnit = action.(battle.PetAction).SuccessControlUnit
			failureControlUnit = action.(battle.PetAction).FailureControlUnit
			successJumpId = action.(battle.PetAction).SuccessJumpId
			failureJumpId = action.(battle.PetAction).FailureJumpId
		}

		if param != "" {
			tag = fmt.Sprintf("%s:%s", tag, param)
		}

		if threshold != "" {
			tag = fmt.Sprintf("%s:%s", tag, threshold)
		}

		if len(successControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(successControlUnit[:1].String())
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%d", tag, successJumpId)
			}
		} else {
			tag = tag + ":"
		}

		if len(failureControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(failureControlUnit[:1].String())
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%d", tag, failureJumpId)
			}
		} else {
			tag = tag + ":"
		}

		tagTextCanvas := canvas.NewText(tag, color.White)
		tagTextCanvas.Alignment = fyne.TextAlignCenter
		tagTextCanvas.TextStyle = fyne.TextStyle{Bold: true, Italic: true, TabWidth: 1}
		tagContainer.Add(tagTextCanvas)
		tagContainers = append(tagContainers, tagContainer)
	}
	return
}
