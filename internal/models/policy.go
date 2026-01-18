package models

type TeamPolicy struct {
	HackathonID  string `json:"hackathon_id"`
	AllowsTeams  bool   `json:"allows_teams"`
	RequiresTeams bool  `json:"requires_teams"`
	MinTeamSize  int    `json:"min_team_size,omitempty"`
	MaxTeamSize  int    `json:"max_team_size,omitempty"`
}

type LeaderboardPolicy struct {
	HackathonID string `json:"hackathon_id"`
	Frozen      bool   `json:"frozen"`
	Published   bool   `json:"published"`
}
