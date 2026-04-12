package repositories

import (
	"database/sql"
	"errors"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/lib"
)

type EnvironmentRepo struct {
	db lib.QueryAble
}

func NewEnvironmentRepo(db lib.QueryAble) ports.EnvironmentRepository {
	return &EnvironmentRepo{db: db}
}

func (r *EnvironmentRepo) ListAll() ([]models.Environment, error) {
	var rows []models.Environment
	err := r.db.Select(&rows, `SELECT id, name, description, created_at, updated_at FROM environments ORDER BY name`)
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return rows, nil
}

func (r *EnvironmentRepo) ListForUser(userID string) ([]models.EnvironmentWithRole, error) {
	var rows []models.EnvironmentWithRole
	err := r.db.Select(&rows, `
SELECT e.id, e.name, e.description, e.created_at, e.updated_at, m.role
FROM environment_members m
INNER JOIN environments e ON e.id = m.environment_id
WHERE m.user_id = $1
ORDER BY e.name
`, userID)
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return rows, nil
}

func (r *EnvironmentRepo) GetByID(id string) (*models.Environment, error) {
	var e models.Environment
	err := r.db.Get(&e, `SELECT id, name, description, created_at, updated_at FROM environments WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &e, nil
}

func (r *EnvironmentRepo) InsertEnvironment(id, name string, description *string, createdAt, updatedAt string) error {
	_, err := r.db.Exec(`
INSERT INTO environments (id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`, id, name, description, createdAt, updatedAt)
	if err != nil {
		return models.CreateErrorWithContext(err)
	}
	return nil
}

func (r *EnvironmentRepo) UpdateEnvironment(id string, updatedAt string, name *string, updateDescription bool, description *string) error {
	if name == nil && !updateDescription {
		return nil
	}
	if name != nil && updateDescription {
		_, err := r.db.Exec(`UPDATE environments SET name = $1, description = $2, updated_at = $3 WHERE id = $4`, *name, description, updatedAt, id)
		return errOrWrap(err)
	}
	if name != nil {
		_, err := r.db.Exec(`UPDATE environments SET name = $1, updated_at = $2 WHERE id = $3`, *name, updatedAt, id)
		return errOrWrap(err)
	}
	_, err := r.db.Exec(`UPDATE environments SET description = $1, updated_at = $2 WHERE id = $3`, description, updatedAt, id)
	return errOrWrap(err)
}

func errOrWrap(err error) error {
	if err == nil {
		return nil
	}
	return models.CreateErrorWithContext(err)
}

func (r *EnvironmentRepo) DeleteEnvironment(id string) error {
	_, err := r.db.Exec(`DELETE FROM environments WHERE id = $1`, id)
	return errOrWrap(err)
}

func (r *EnvironmentRepo) GetMembership(environmentID, userID string) (*models.EnvironmentMember, error) {
	var m models.EnvironmentMember
	err := r.db.Get(&m, `
SELECT environment_id, user_id, role FROM environment_members
WHERE environment_id = $1 AND user_id = $2
`, environmentID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &m, nil
}

func (r *EnvironmentRepo) ListMembers(environmentID string) ([]models.EnvironmentMemberRow, error) {
	var rows []models.EnvironmentMemberRow
	err := r.db.Select(&rows, `
SELECT m.user_id, m.role, u.username, u.auth_uuid
FROM environment_members m
INNER JOIN hive_users u ON u.id = m.user_id
WHERE m.environment_id = $1
ORDER BY u.username
`, environmentID)
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return rows, nil
}

func (r *EnvironmentRepo) UpsertMember(environmentID, userID string, role models.MemberRole) error {
	_, err := r.db.Exec(`
INSERT INTO environment_members (environment_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (environment_id, user_id) DO UPDATE SET role = EXCLUDED.role
`, environmentID, userID, string(role))
	return errOrWrap(err)
}

func (r *EnvironmentRepo) InsertMember(environmentID, userID string, role models.MemberRole) error {
	return r.UpsertMember(environmentID, userID, role)
}

func (r *EnvironmentRepo) RemoveMember(environmentID, userID string) error {
	_, err := r.db.Exec(`DELETE FROM environment_members WHERE environment_id = $1 AND user_id = $2`, environmentID, userID)
	return errOrWrap(err)
}
