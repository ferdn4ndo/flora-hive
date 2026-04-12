package mqtttopic

import (
	"strings"
)

// ParseDeviceRowIDFromTopic returns the first segment after optional prefix (catalog devices.id).
func ParseDeviceRowIDFromTopic(topic, topicPrefix string) string {
	t := strings.TrimLeft(topic, "/")
	p := strings.TrimSuffix(strings.TrimSpace(topicPrefix), "/")
	if p != "" && strings.HasPrefix(t, p+"/") {
		t = t[len(p)+1:]
	}
	seg := strings.SplitN(t, "/", 2)[0]
	return seg
}

// ExtractWildcardSegments matches a pattern with '+' single-level wildcards to a topic.
func ExtractWildcardSegments(pattern, topic string) []string {
	pa := strings.Split(pattern, "/")
	ta := strings.Split(topic, "/")
	if len(pa) != len(ta) {
		return nil
	}
	segments := make([]string, 0, len(pa))
	for i := range pa {
		if pa[i] == "+" {
			segments = append(segments, ta[i])
		} else if pa[i] != ta[i] {
			return nil
		}
	}
	if len(segments) == 0 {
		return nil
	}
	return segments
}

func CompositeDeviceID(segments []string) string {
	if len(segments) == 1 {
		return segments[0]
	}
	return strings.Join(segments, "/")
}

// HiveIdentityFromSegments returns deviceRowId when the pattern yields a single segment (row UUID).
func HiveIdentityFromSegments(segments []string) (deviceRowID string, ok bool) {
	if len(segments) != 1 {
		return "", false
	}
	return segments[0], true
}

// ErrEmptyTopic is returned for invalid publish topics.
type ErrEmptyTopic struct{}

func (e ErrEmptyTopic) Error() string {
	return "topic is empty"
}

// NormalizePublishTopic applies optional topic prefix for publish paths.
func NormalizePublishTopic(topic, topicPrefix string) (string, error) {
	t := strings.TrimSpace(topic)
	t = strings.TrimPrefix(t, "/")
	if t == "" {
		return "", ErrEmptyTopic{}
	}
	p := strings.TrimSuffix(strings.TrimSpace(topicPrefix), "/")
	if p != "" && t != p && !strings.HasPrefix(t, p+"/") {
		t = p + "/" + t
	}
	return t, nil
}
