package container

import (
	"cg/game/battle"
	"cg/game/enum/controlunit"
	"cg/game/enum/role"
	"cg/game/enum/threshold"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const tagMinimumWidth float32 = 48

func generateTags(actionState battle.ActionState) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainers(actionState, role.Character)
	tagContaines = append(tagContaines, createTagContainers(actionState, role.Pet)...)
	return
}

var (
	characterFinishingTagColor   = color.RGBA{0x33, 0x4E, 0x9A, 0xFF}
	characterConditionalTagColor = color.RGBA{0x00, 0x6B, 0x78, 0xFF}
	characterSpecialTagColor     = color.RGBA{0x45, 0x5A, 0x64, 0xFF}
	petFinishingTagColor         = color.RGBA{0x9A, 0x34, 0x12, 0xFF}
	petConditionalTagColor       = color.RGBA{0x8A, 0x5A, 0x00, 0xFF}
	petSpecialTagColor           = color.RGBA{0x7A, 0x3E, 0x5D, 0xFF}
)

func newTagContainer(tag string, tagColor color.Color) *fyne.Container {
	tagCanvas := canvas.NewRectangle(tagColor)
	tagCanvas.SetMinSize(fyne.NewSize(tagMinimumWidth, 0))

	tagTextCanvas := canvas.NewText(tag, color.White)
	tagTextCanvas.Alignment = fyne.TextAlignCenter
	tagTextCanvas.TextStyle = fyne.TextStyle{Bold: true, Italic: true, TabWidth: 1}

	return container.NewStack(tagCanvas, container.NewPadded(tagTextCanvas))
}

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

		tagContainers = append(tagContainers, newTagContainer(tag, tagColor))
	}
	return
}
