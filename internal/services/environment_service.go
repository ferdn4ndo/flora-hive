package services

import (
	"time"

	"github.com/google/uuid"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/internal/infrastructure/repositories"
	"flora-hive/lib"
)

// EnvironmentService contains environment and membership operations.
type EnvironmentService struct {
	handler ports.DatabaseHandler
	env     ports.EnvironmentRepository
}

// NewEnvironmentService constructs the service.
func NewEnvironmentService(handler ports.DatabaseHandler, db lib.Database) *EnvironmentService {
	return &EnvironmentService{
		handler: handler,
		env:     repositories.NewEnvironmentRepo(db),
	}
}

// ListAll returns every environment.
func (s *EnvironmentService) ListAll() ([]models.Environment, error) {
	return s.env.ListAll()
}

// ListForUser returns environments visible to a user with roles.
func (s *EnvironmentService) ListForUser(userID string) ([]models.EnvironmentWithRole, error) {
	return s.env.ListForUser(userID)
}

// GetByID fetches an environment by id.
func (s *EnvironmentService) GetByID(id string) (*models.Environment, error) {
	return s.env.GetByID(id)
}

// Create creates an environment and makes the creator an editor.
func (s *EnvironmentService) Create(name string, description *string, creatorUserID string) (*models.Environment, error) {
	id := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var out *models.Environment
	err := s.handler.WithTrx(func(r *ports.Repositories) error {
		if err := r.Environment.InsertEnvironment(id, name, description, now, now); err != nil {
			return err
		}
		if err := r.Environment.InsertMember(id, creatorUserID, models.RoleEditor); err != nil {
			return err
		}
		var err error
		out, err = r.Environment.GetByID(id)
		return err
	})
	return out, err
}

// Update patches name/description.
func (s *EnvironmentService) Update(id string, name *string, updateDescription bool, description *string) (*models.Environment, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err := s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Environment.UpdateEnvironment(id, now, name, updateDescription, description)
	})
	if err != nil {
		return nil, err
	}
	return s.env.GetByID(id)
}

// Delete removes an environment.
func (s *EnvironmentService) Delete(id string) error {
	return s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Environment.DeleteEnvironment(id)
	})
}

// GetMembership returns membership for a user in an environment.
func (s *EnvironmentService) GetMembership(environmentID, userID string) (*models.EnvironmentMember, error) {
	return s.env.GetMembership(environmentID, userID)
}

// RequireEnvAccess enforces RBAC; needWrite requires editor.
func (s *EnvironmentService) RequireEnvAccess(environmentID, userID string, needWrite bool) (*models.EnvironmentMember, error) {
	m, err := s.env.GetMembership(environmentID, userID)
	if err != nil {
		return nil, err
	}
	if m == nil || !CanReadRole(m.Role) {
		return nil, &models.ForbiddenError{Message: "No access to this environment"}
	}
	if needWrite && !CanWriteRole(m.Role) {
		return nil, &models.ForbiddenError{Message: "Editor role required"}
	}
	return m, nil
}

// ListMembers lists members with usernames.
func (s *EnvironmentService) ListMembers(environmentID string) ([]models.EnvironmentMemberRow, error) {
	return s.env.ListMembers(environmentID)
}

// UpsertMember adds or updates a member.
func (s *EnvironmentService) UpsertMember(environmentID, userID string, role models.MemberRole) error {
	return s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Environment.UpsertMember(environmentID, userID, role)
	})
}

// RemoveMember removes a member.
func (s *EnvironmentService) RemoveMember(environmentID, userID string) error {
	return s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Environment.RemoveMember(environmentID, userID)
	})
}
