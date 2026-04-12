package models

// HiveUser maps to hive_users.
type HiveUser struct {
	ID         string `db:"id"`
	AuthUUID   string `db:"auth_uuid"`
	Username   string `db:"username"`
	SystemName string `db:"system_name"`
	UpdatedAt  string `db:"updated_at"`
}

// Environment maps to environments.
type Environment struct {
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
	CreatedAt   string  `db:"created_at"`
	UpdatedAt   string  `db:"updated_at"`
}

// MemberRole is viewer or editor.
type MemberRole string

const (
	RoleViewer MemberRole = "viewer"
	RoleEditor MemberRole = "editor"
)

// EnvironmentMember maps to environment_members.
type EnvironmentMember struct {
	EnvironmentID string     `db:"environment_id"`
	UserID        string     `db:"user_id"`
	Role          MemberRole `db:"role"`
}

// Device maps to devices.
type Device struct {
	ID             string  `db:"id"`
	EnvironmentID  string  `db:"environment_id"`
	ParentDeviceID *string `db:"parent_device_id"`
	DeviceType     string  `db:"device_type"`
	DeviceID       string  `db:"device_id"`
	DisplayName    *string `db:"display_name"`
	CreatedAt      string  `db:"created_at"`
	UpdatedAt      string  `db:"updated_at"`
}

// EnvironmentWithRole is a joined row for listing environments for a user.
type EnvironmentWithRole struct {
	Environment
	Role MemberRole `db:"role"`
}

// EnvironmentMemberRow is a member listing with user profile fields.
type EnvironmentMemberRow struct {
	UserID   string `json:"userId" db:"user_id"`
	Role     string `json:"role" db:"role"`
	Username string `json:"username" db:"username"`
	AuthUUID string `json:"authUuid" db:"auth_uuid"`
}
