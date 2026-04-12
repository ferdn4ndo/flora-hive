package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/lib"
)

type DeviceRepo struct {
	db lib.QueryAble
}

func NewDeviceRepo(db lib.QueryAble) ports.DeviceRepository {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) ListByEnvironment(environmentID string, parent *ports.ParentDeviceFilter) ([]models.Device, error) {
	q := `SELECT id, environment_id, parent_device_id, device_type, device_id, display_name, created_at, updated_at FROM devices WHERE environment_id = $1`
	args := []interface{}{environmentID}
	if parent != nil {
		if parent.RootOnly {
			q += ` AND parent_device_id IS NULL`
		} else if !parent.All && parent.ID != "" {
			q += ` AND parent_device_id = $2`
			args = append(args, parent.ID)
		}
	}
	q += ` ORDER BY device_id`
	var rows []models.Device
	err := r.db.Select(&rows, q, args...)
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return rows, nil
}

func (r *DeviceRepo) ListRowIDsForEnvironments(envIDs []string) ([]string, error) {
	if len(envIDs) == 0 {
		return nil, nil
	}
	var ids []string
	err := r.db.Select(&ids, `SELECT id FROM devices WHERE environment_id = ANY($1)`, pq.Array(envIDs))
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return ids, nil
}

func (r *DeviceRepo) Insert(d *models.Device) error {
	_, err := r.db.Exec(`
INSERT INTO devices (id, environment_id, parent_device_id, device_type, device_id, display_name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, d.ID, d.EnvironmentID, d.ParentDeviceID, d.DeviceType, d.DeviceID, d.DisplayName, d.CreatedAt, d.UpdatedAt)
	return errOrWrap(err)
}

func (r *DeviceRepo) GetByID(id string) (*models.Device, error) {
	var d models.Device
	err := r.db.Get(&d, `SELECT id, environment_id, parent_device_id, device_type, device_id, display_name, created_at, updated_at FROM devices WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &d, nil
}

func (r *DeviceRepo) GetByEnvAndDeviceID(environmentID, logicalDeviceID string) (*models.Device, error) {
	var d models.Device
	err := r.db.Get(&d, `
SELECT id, environment_id, parent_device_id, device_type, device_id, display_name, created_at, updated_at
FROM devices WHERE environment_id = $1 AND device_id = $2
`, environmentID, logicalDeviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.CreateErrorWithContext(err)
	}
	return &d, nil
}

func (r *DeviceRepo) ListByLogicalDeviceIDGlobally(logicalDeviceID string) ([]models.Device, error) {
	var rows []models.Device
	err := r.db.Select(&rows, `
SELECT id, environment_id, parent_device_id, device_type, device_id, display_name, created_at, updated_at
FROM devices WHERE device_id = $1
ORDER BY environment_id, id
`, logicalDeviceID)
	if err != nil {
		return nil, models.CreateErrorWithContext(err)
	}
	return rows, nil
}

func (r *DeviceRepo) Update(id string, updatedAt string, deviceType, logicalDeviceID *string, updateDisplayName bool, displayName *string, updateParent bool, parentDeviceID *string) error {
	parts := make([]string, 0, 5)
	args := make([]interface{}, 0, 8)
	if deviceType != nil {
		parts = append(parts, fmt.Sprintf("device_type = $%d", len(args)+1))
		args = append(args, *deviceType)
	}
	if logicalDeviceID != nil {
		parts = append(parts, fmt.Sprintf("device_id = $%d", len(args)+1))
		args = append(args, *logicalDeviceID)
	}
	if updateDisplayName {
		parts = append(parts, fmt.Sprintf("display_name = $%d", len(args)+1))
		args = append(args, displayName)
	}
	if updateParent {
		parts = append(parts, fmt.Sprintf("parent_device_id = $%d", len(args)+1))
		args = append(args, parentDeviceID)
	}
	if len(parts) == 0 {
		return nil
	}
	parts = append(parts, fmt.Sprintf("updated_at = $%d", len(args)+1))
	args = append(args, updatedAt)
	partsJoin := strings.Join(parts, ", ")
	args = append(args, id)
	q := fmt.Sprintf(`UPDATE devices SET %s WHERE id = $%d`, partsJoin, len(args))
	_, err := r.db.Exec(q, args...)
	return errOrWrap(err)
}

func (r *DeviceRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM devices WHERE id = $1`, id)
	return errOrWrap(err)
}
