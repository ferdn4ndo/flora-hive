package lib

import (
	"reflect"
	"testing"
)

func TestParseHiveAPIKeysFromRaw(t *testing.T) {
	t.Parallel()
	got := ParseHiveAPIKeysFromRaw(" a , b , ")
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if ParseHiveAPIKeysFromRaw("") != nil {
		t.Fatal("expected nil for empty")
	}
	if ParseHiveAPIKeysFromRaw("   ") != nil {
		t.Fatal("expected nil for whitespace-only")
	}
}
