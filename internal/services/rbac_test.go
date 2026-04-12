package services

import (
	"testing"

	"flora-hive/internal/domain/models"
)

func TestCanReadRole(t *testing.T) {
	t.Parallel()
	if !CanReadRole(models.RoleViewer) {
		t.Fatal("viewer should read")
	}
	if !CanReadRole(models.RoleEditor) {
		t.Fatal("editor should read")
	}
	if CanReadRole("") {
		t.Fatal("empty role should not read")
	}
	if CanReadRole(models.MemberRole("admin")) {
		t.Fatal("unknown role should not read")
	}
}

func TestCanWriteRole(t *testing.T) {
	t.Parallel()
	if CanWriteRole(models.RoleViewer) {
		t.Fatal("viewer should not write")
	}
	if !CanWriteRole(models.RoleEditor) {
		t.Fatal("editor should write")
	}
	if CanWriteRole("") {
		t.Fatal("empty role should not write")
	}
	if CanWriteRole(models.MemberRole("admin")) {
		t.Fatal("unknown role should not write")
	}
}
