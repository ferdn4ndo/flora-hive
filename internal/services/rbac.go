package services

import "flora-hive/internal/domain/models"

// CanReadRole returns true for viewer and editor.
func CanReadRole(role models.MemberRole) bool {
	return role == models.RoleViewer || role == models.RoleEditor
}

// CanWriteRole returns true only for editor.
func CanWriteRole(role models.MemberRole) bool {
	return role == models.RoleEditor
}
