package enum_test

import (
	"testing"

	"cg/game/enum/character"
	"cg/game/enum/movement"
	"cg/game/enum/pet"
)

func TestUserFacingEnumLabels(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "biased diagonal movement", got: string(movement.BIASED_DIAGONAL), want: "Biased Diagonal"},
		{name: "biased reversed diagonal movement", got: string(movement.BIASED_REVERSED_DIAGONAL), want: "Biased Reversed Diagonal"},
		{name: "character wait", got: character.Hang.String(), want: "Wait"},
		{name: "character heal ally", got: character.HealOne.String(), want: "*Heal Ally"},
		{name: "character heal party", got: character.HealMulti.String(), want: "*Heal Party"},
		{name: "character T-shaped skill", got: character.TShapedSkill.String(), want: "*T-Shaped Skill"},
		{name: "pet wait", got: pet.Hang.String(), want: "Pet Wait"},
		{name: "pet heal ally", got: pet.HealOne.String(), want: "*Pet Heal Ally"},
		{name: "pet dismount", got: pet.OffRide.String(), want: "*Pet Dismount"},
		{name: "pet T-shaped skill", got: pet.TShapedSkill.String(), want: "*Pet T-Shaped Skill"},
		{name: "pet threshold skill", got: pet.ThresholdSkill.String(), want: "*Pet Threshold Skill"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("label = %q, want %q", test.got, test.want)
			}
		})
	}
}

func TestActionEnumValues(t *testing.T) {
	if character.TrainSkill != 17 || character.TShapedSkill != 18 {
		t.Fatalf("character action values = (TrainSkill: %d, TShapedSkill: %d), want (17, 18)", character.TrainSkill, character.TShapedSkill)
	}
	if pet.Hang != 9 || pet.TShapedSkill != 10 || pet.ThresholdSkill != 11 {
		t.Fatalf("pet action values = (Hang: %d, TShapedSkill: %d, ThresholdSkill: %d), want (9, 10, 11)", pet.Hang, pet.TShapedSkill, pet.ThresholdSkill)
	}
}
