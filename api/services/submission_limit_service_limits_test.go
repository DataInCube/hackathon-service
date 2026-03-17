package services

import (
	"errors"
	"testing"
)

func TestValidateSubmissionLimits(t *testing.T) {
	if err := validateSubmissionLimits(1, 10, 3); err != nil {
		t.Fatalf("expected valid submission limits, got=%v", err)
	}

	cases := [][3]int{
		{-1, 10, 3},
		{1, -10, 3},
		{1, 10, -3},
	}
	for i, tc := range cases {
		err := validateSubmissionLimits(tc[0], tc[1], tc[2])
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("case %d: expected ErrInvalid, got=%v", i, err)
		}
	}
}
