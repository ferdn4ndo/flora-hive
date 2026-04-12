package mqtttopic

import (
	"testing"
)

func TestNormalizePublishTopic(t *testing.T) {
	t.Parallel()
	t.Run("prefixes relative topic", func(t *testing.T) {
		t.Parallel()
		got, err := NormalizePublishTopic("dev/heartbeat", "flora")
		if err != nil {
			t.Fatal(err)
		}
		if want := "flora/dev/heartbeat"; got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
	t.Run("leaves already prefixed topic", func(t *testing.T) {
		t.Parallel()
		got, err := NormalizePublishTopic("flora/lab/d1/cmd", "flora")
		if err != nil {
			t.Fatal(err)
		}
		if want := "flora/lab/d1/cmd"; got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
	t.Run("empty topic errors", func(t *testing.T) {
		t.Parallel()
		_, err := NormalizePublishTopic("   ", "flora")
		if err == nil {
			t.Fatal("expected error for empty topic")
		}
	})
}

func TestParseDeviceRowIDFromTopic(t *testing.T) {
	t.Parallel()
	if got := ParseDeviceRowIDFromTopic("flora/abc-uuid/heartbeat", "flora"); got != "abc-uuid" {
		t.Fatalf("got %q", got)
	}
	if got := ParseDeviceRowIDFromTopic("abc-uuid/heartbeat", ""); got != "abc-uuid" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractWildcardSegments(t *testing.T) {
	t.Parallel()
	pat := "flora/+/heartbeat"
	topic := "flora/device-1/heartbeat"
	segs := ExtractWildcardSegments(pat, topic)
	if len(segs) != 1 || segs[0] != "device-1" {
		t.Fatalf("got %#v", segs)
	}
	if ExtractWildcardSegments(pat, "other/a/heartbeat") != nil {
		t.Fatal("expected nil")
	}
}
