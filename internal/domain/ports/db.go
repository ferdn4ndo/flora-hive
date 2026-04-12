package ports

import "flora-hive/internal/domain/models"

// Repositories groups transactional data access.
type Repositories struct {
	HiveUser    HiveUserRepository
	Environment EnvironmentRepository
	Device      DeviceRepository
}

// DatabaseHandler runs work inside a transaction.
type DatabaseHandler interface {
	WithTrx(cb func(r *Repositories) error) error
}

// HiveUserRepository persists hive-linked users.
type HiveUserRepository interface {
	GetByAuthUUID(authUUID string) (*models.HiveUser, error)
	UpsertFromMe(authUUID, username, systemName, now string) (id string, err error)
	GetByID(id string) (*models.HiveUser, error)
}

// EnvironmentRepository manages environments and membership.
type EnvironmentRepository interface {
	ListAll() ([]models.Environment, error)
	ListForUser(userID string) ([]models.EnvironmentWithRole, error)
	GetByID(id string) (*models.Environment, error)
	InsertEnvironment(id, name string, description *string, createdAt, updatedAt string) error
	// UpdateEnvironment sets name when name != nil. When updateDescription is true, description is set (NULL when description is nil).
	UpdateEnvironment(id string, updatedAt string, name *string, updateDescription bool, description *string) error
	DeleteEnvironment(id string) error
	GetMembership(environmentID, userID string) (*models.EnvironmentMember, error)
	ListMembers(environmentID string) ([]models.EnvironmentMemberRow, error)
	UpsertMember(environmentID, userID string, role models.MemberRole) error
	RemoveMember(environmentID, userID string) error
	InsertMember(environmentID, userID string, role models.MemberRole) error
}

// DeviceRepository manages devices per environment.
type DeviceRepository interface {
	ListByEnvironment(environmentID string, parent *ParentDeviceFilter) ([]models.Device, error)
	ListRowIDsForEnvironments(envIDs []string) ([]string, error)
	Insert(d *models.Device) error
	GetByID(id string) (*models.Device, error)
	GetByEnvAndDeviceID(environmentID, logicalDeviceID string) (*models.Device, error)
	// ListByLogicalDeviceIDGlobally returns all device rows with the given devices.device_id (may be >1 across environments).
	ListByLogicalDeviceIDGlobally(logicalDeviceID string) ([]models.Device, error)
	Update(id string, updatedAt string, deviceType, logicalDeviceID *string, updateDisplayName bool, displayName *string, updateParent bool, parentDeviceID *string) error
	Delete(id string) error
}

// ParentDeviceFilter selects devices by parent_device_id semantics.
type ParentDeviceFilter struct {
	// All omits parent filter; RootOnly filters parent_device_id IS NULL; ID sets parent_device_id = ID
	All      bool
	RootOnly bool
	ID       string
}

// NewParentFilterAll lists all devices in the environment.
func NewParentFilterAll() *ParentDeviceFilter {
	return &ParentDeviceFilter{All: true}
}

// NewParentFilterRoot lists only root devices (no parent).
func NewParentFilterRoot() *ParentDeviceFilter {
	return &ParentDeviceFilter{RootOnly: true}
}

// NewParentFilterID lists devices with a specific parent row id.
func NewParentFilterID(parentID string) *ParentDeviceFilter {
	return &ParentDeviceFilter{ID: parentID}
}
