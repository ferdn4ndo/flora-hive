package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/ports"
	"flora-hive/internal/infrastructure/repositories"
	"flora-hive/lib"
)

// ErrInvalidParent is returned when parent_device_id does not belong to the environment.
var ErrInvalidParent = errors.New("invalid parent device")

// DeviceService manages devices within environments.
type DeviceService struct {
	handler ports.DatabaseHandler
	dev     ports.DeviceRepository
	es      *EnvironmentService
}

// NewDeviceService constructs the service.
func NewDeviceService(handler ports.DatabaseHandler, db lib.Database, es *EnvironmentService) *DeviceService {
	return &DeviceService{
		handler: handler,
		dev:     repositories.NewDeviceRepo(db),
		es:      es,
	}
}

// ListByEnvironment lists devices with optional parent filter (nil = all).
func (s *DeviceService) ListByEnvironment(environmentID string, parent *ports.ParentDeviceFilter) ([]models.Device, error) {
	return s.dev.ListByEnvironment(environmentID, parent)
}

// ListInEnvironment enforces read access then lists devices.
func (s *DeviceService) ListInEnvironment(environmentID, userID string, parent *ports.ParentDeviceFilter) ([]models.Device, error) {
	if _, err := s.es.RequireEnvAccess(environmentID, userID, false); err != nil {
		return nil, err
	}
	return s.dev.ListByEnvironment(environmentID, parent)
}

// ListRowIDsForEnvironments returns catalog device row ids for MQTT filtering.
func (s *DeviceService) ListRowIDsForEnvironments(envIDs []string) ([]string, error) {
	return s.dev.ListRowIDsForEnvironments(envIDs)
}

// CreateDevice creates a device after editor check and parent validation.
func (s *DeviceService) CreateDevice(environmentID, userID, deviceType, logicalDeviceID string, displayName *string, parentDeviceID *string) (*models.Device, error) {
	if _, err := s.es.RequireEnvAccess(environmentID, userID, true); err != nil {
		return nil, err
	}
	if parentDeviceID != nil && *parentDeviceID != "" {
		p, err := s.dev.GetByID(*parentDeviceID)
		if err != nil {
			return nil, err
		}
		if p == nil || p.EnvironmentID != environmentID {
			return nil, ErrInvalidParent
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := uuid.New().String()
	d := &models.Device{
		ID:             id,
		EnvironmentID:  environmentID,
		ParentDeviceID: parentDeviceID,
		DeviceType:     deviceType,
		DeviceID:       logicalDeviceID,
		DisplayName:    displayName,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err := s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Device.Insert(d)
	})
	if err != nil {
		return nil, err
	}
	return s.dev.GetByID(id)
}

// GetRowByID returns a device row without access check (internal).
func (s *DeviceService) GetRowByID(deviceRowID string) (*models.Device, error) {
	return s.dev.GetByID(deviceRowID)
}

// GetByEnvAndDeviceID enforces read access.
func (s *DeviceService) GetByEnvAndDeviceID(environmentID, logicalDeviceID, userID string) (*models.Device, error) {
	d, err := s.dev.GetByEnvAndDeviceID(environmentID, logicalDeviceID)
	if err != nil || d == nil {
		return d, err
	}
	if _, err := s.es.RequireEnvAccess(d.EnvironmentID, userID, false); err != nil {
		return nil, err
	}
	return d, nil
}

// GetRowByEnvAndDeviceID returns a row without membership check (for API key paths).
func (s *DeviceService) GetRowByEnvAndDeviceID(environmentID, logicalDeviceID string) (*models.Device, error) {
	return s.dev.GetByEnvAndDeviceID(environmentID, logicalDeviceID)
}

// UpdateByEnvAndDeviceID updates with editor permission.
func (s *DeviceService) UpdateByEnvAndDeviceID(environmentID, logicalDeviceID, userID string, deviceType, logicalID *string, updateDisplay bool, displayName *string, updateParent bool, parentDeviceID *string) (*models.Device, error) {
	d, err := s.dev.GetByEnvAndDeviceID(environmentID, logicalDeviceID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, nil
	}
	if _, err := s.es.RequireEnvAccess(d.EnvironmentID, userID, true); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err = s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Device.Update(d.ID, now, deviceType, logicalID, updateDisplay, displayName, updateParent, parentDeviceID)
	})
	if err != nil {
		return nil, err
	}
	return s.dev.GetByID(d.ID)
}

// DeleteByEnvAndDeviceID deletes with editor permission.
func (s *DeviceService) DeleteByEnvAndDeviceID(environmentID, logicalDeviceID, userID string) (bool, error) {
	d, err := s.dev.GetByEnvAndDeviceID(environmentID, logicalDeviceID)
	if err != nil {
		return false, err
	}
	if d == nil {
		return false, nil
	}
	if _, err := s.es.RequireEnvAccess(d.EnvironmentID, userID, true); err != nil {
		return false, err
	}
	if err := s.handler.WithTrx(func(r *ports.Repositories) error {
		return r.Device.Delete(d.ID)
	}); err != nil {
		return false, err
	}
	return true, nil
}
