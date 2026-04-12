package repositories

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/lib"
)

type HiveUserRepo struct {
	db lib.QueryAble
}

func NewHiveUserRepo(db lib.QueryAble) ports.HiveUserRepository {
	return &HiveUserRepo{db: db}
}

func (r *HiveUserRepo) GetByAuthUUID(authUUID string) (*models.HiveUser, error) {
	var u models.HiveUser
	err := r.db.Get(&u, `SELECT id, auth_uuid, username, system_name, updated_at FROM hive_users WHERE auth_uuid = $1`, authUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &u, nil
}

func (r *HiveUserRepo) GetByID(id string) (*models.HiveUser, error) {
	var u models.HiveUser
	err := r.db.Get(&u, `SELECT id, auth_uuid, username, system_name, updated_at FROM hive_users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &u, nil
}

func (r *HiveUserRepo) UpsertFromMe(authUUID, username, systemName, now string) (string, error) {
	newID := uuid.New().String()
	var id string
	err := r.db.Get(&id, `
INSERT INTO hive_users (id, auth_uuid, username, system_name, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (auth_uuid) DO UPDATE SET
  username = EXCLUDED.username,
  system_name = EXCLUDED.system_name,
  updated_at = EXCLUDED.updated_at
RETURNING id
`, newID, authUUID, username, systemName, now)
	if err != nil {
		return "", models.CreateErrorWithContext(err)
	}
	return id, nil
}
