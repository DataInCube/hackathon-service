package services

import (
	"testing"
	"time"

	"github.com/DataInCube/hackathon-service/internal/models"
)

func TestValidateHackathonInput(t *testing.T) {
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	cases := []struct {
		name    string
		input   models.Hackathon
		expects bool
	}{
		{
			name: "missing title",
			input: models.Hackathon{Title: ""},
		},
		{
			name: "end before start",
			input: models.Hackathon{Title: "Test", StartsAt: &end, EndsAt: &start},
		},
		{
			name: "requires teams without allows teams",
			input: models.Hackathon{Title: "Test", RequiresTeams: true, AllowsTeams: false},
		},
		{
			name: "negative team sizes",
			input: models.Hackathon{Title: "Test", MinTeamSize: -1},
		},
		{
			name: "min greater than max",
			input: models.Hackathon{Title: "Test", MinTeamSize: 5, MaxTeamSize: 3},
		},
		{
			name: "valid",
			input: models.Hackathon{Title: "Test", StartsAt: &start, EndsAt: &end, AllowsTeams: true, RequiresTeams: true, MinTeamSize: 2, MaxTeamSize: 5},
			expects: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateHackathonInput(tc.input)
			if tc.expects && err != nil {
				t.Fatalf("expected valid input, got error: %v", err)
			}
			if !tc.expects && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestIsTransitionAllowed(t *testing.T) {
	allowed := []struct {
		from string
		to   string
	}{
		{models.HackathonStateDraft, models.HackathonStatePublished},
		{models.HackathonStatePublished, models.HackathonStateWarmup},
		{models.HackathonStateWarmup, models.HackathonStateLive},
		{models.HackathonStateLive, models.HackathonStateSubmissionFrozen},
		{models.HackathonStateSubmissionFrozen, models.HackathonStateEvaluationOnly},
		{models.HackathonStateEvaluationOnly, models.HackathonStateCompleted},
		{models.HackathonStateCompleted, models.HackathonStateArchived},
	}

	for _, tc := range allowed {
		if !isTransitionAllowed(tc.from, tc.to) {
			t.Fatalf("expected transition allowed from %s to %s", tc.from, tc.to)
		}
	}

	if isTransitionAllowed(models.HackathonStateDraft, models.HackathonStateLive) {
		t.Fatalf("expected transition from draft to live to be disallowed")
	}
}

func TestIsStateAtLeast(t *testing.T) {
	if !isStateAtLeast(models.HackathonStateLive, models.HackathonStateWarmup) {
		t.Fatalf("expected live to be at least warmup")
	}
	if isStateAtLeast(models.HackathonStatePublished, models.HackathonStateLive) {
		t.Fatalf("expected published to be before live")
	}
}
