package service

import(
	"context"
	"database/sql"

	"github.com/DataInCube/hackathon-service/internal/models"
)

type RegistrationService interface {
	NewRegistrationService(db sql.DB) (RegistrationService)

	Register(ctx context.Context, r models.Registration) (int64, error)

	GetAllRegistrations(ctx context.Context) ([]models.Registration, error)

	GetRegistrationByID(ctx context.Context, id uint) (*models.Registration, error)

	UpdateRegistration(ctx context.Context, id uint, r models.Registration) error

	DeleteRegistration(ctx context.Context, id uint) error

	// Individual registration
	RegisterIndividual(ctx context.Context, participantID, hackathonID uint) (uint, error)

	// Register participant into an existing team
	RegisterToTeam(ctx context.Context, participantID, hackathonID, teamID uint) (uint, error)

	// Approve a participant’s join request (by team lead)
	ApproveTeamJoin(ctx context.Context, teamLeadID, participantID, teamID uint) error

	// Get all registrations for a hackathon
	GetRegistrationsByHackathon(ctx context.Context, hackathonID uint) ([]models.Registration, error)

	// Get participant’s registration
	GetRegistrationByParticipant(ctx context.Context, participantID, hackathonID uint) (*models.Registration, error)
}
