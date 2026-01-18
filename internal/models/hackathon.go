package models

import (
	"encoding/json"
	"time"
)

type Hackathon struct {
	ID                    string          `json:"id"`
	Title                 string          `json:"title"`
	Description           string          `json:"description"`
	State                 string          `json:"state"`
	Visibility            string          `json:"visibility"`
	StartsAt              *time.Time      `json:"starts_at,omitempty"`
	EndsAt                *time.Time      `json:"ends_at,omitempty"`
	AllowsTeams           bool            `json:"allows_teams"`
	RequiresTeams         bool            `json:"requires_teams"`
	MinTeamSize           int             `json:"min_team_size,omitempty"`
	MaxTeamSize           int             `json:"max_team_size,omitempty"`
	ActiveRuleVersionID   *string         `json:"active_rule_version_id,omitempty"`
	LeaderboardFrozen     bool            `json:"leaderboard_frozen"`
	LeaderboardPublished  bool            `json:"leaderboard_published"`
	CreatedBy             string          `json:"created_by,omitempty"`
	Metadata              json.RawMessage `json:"metadata,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	PublishedAt           *time.Time      `json:"published_at,omitempty"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty"`
	ArchivedAt            *time.Time      `json:"archived_at,omitempty"`
}
