package services

import (
	"time"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/internal/infrastructure/repositories"
	"flora-hive/internal/infrastructure/userver"
	"flora-hive/lib"
)

// UserService syncs hive_users from uServer /me payloads.
type UserService struct {
	users ports.HiveUserRepository
}

// NewUserService constructs the service.
func NewUserService(db lib.Database) *UserService {
	return &UserService{users: repositories.NewHiveUserRepo(db)}
}

// UpsertFromMe inserts or updates a hive user and returns hive user id.
func (s *UserService) UpsertFromMe(me *userver.MeResponse) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return s.users.UpsertFromMe(me.UUID, me.Username, me.SystemName, now)
}

// FindByAuthUUID returns a hive user by uServer auth uuid.
func (s *UserService) FindByAuthUUID(authUUID string) (*models.HiveUser, error) {
	return s.users.GetByAuthUUID(authUUID)
}
