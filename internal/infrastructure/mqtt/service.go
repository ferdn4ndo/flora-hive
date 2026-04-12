package mqtt

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/fx"

	"flora-hive/internal/domain/mqtttopic"
	"flora-hive/internal/domain/ports"
	"flora-hive/internal/infrastructure/repositories"
	"flora-hive/lib"
)

// PresenceKind distinguishes TTL-based vs explicit connection signals.
type PresenceKind string

const (
	PresenceTTL      PresenceKind = "ttl"
	PresenceExplicit PresenceKind = "explicit"
)

// RegistryEntry tracks last seen device state from MQTT.
type RegistryEntry struct {
	ID                string
	LastSeenAt        string
	LastTopic         string
	PresenceKind      PresenceKind
	ExplicitConnected *bool
	DeviceRowID       string // catalog devices.id (resolved from devices.device_id in topic)
	Meta              interface{}
}

// Identity links live MQTT traffic to a catalog row id.
type Identity struct {
	DeviceRowID string `json:"deviceRowId"`
}

// PublicDevice is returned by the API.
type PublicDevice struct {
	ID         string      `json:"id"`
	Connected  bool        `json:"connected"`
	LastSeenAt string      `json:"lastSeenAt"`
	LastTopic  string      `json:"lastTopic"`
	Identity   *Identity   `json:"identity,omitempty"`
	Telemetry  interface{} `json:"telemetry,omitempty"`
}

// State is connection info for GET /mqtt/connection.
type State struct {
	Connected bool    `json:"connected"`
	ClientID  string  `json:"clientId"`
	URL       string  `json:"url"`
	LastError *string `json:"lastError"`
}

// Service manages the MQTT client and in-memory device registry.
type Service struct {
	env    lib.Env
	logger lib.Logger
	dev    ports.DeviceRepository

	mu          sync.RWMutex
	client      mqtt.Client
	connected   bool
	lastError   string
	registry    map[string]*RegistryEntry
	subscribeMu sync.Mutex
}

// NewService constructs the service (client connects on fx OnStart).
func NewService(env lib.Env, logger lib.Logger, db lib.Database, lc fx.Lifecycle) *Service {
	rp := repositories.NewDeviceRepo(&db)
	s := &Service{env: env, logger: logger, dev: rp, registry: make(map[string]*RegistryEntry)}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.Connect()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.Disconnect()
			return nil
		},
	})
	return s
}

func redactURL(raw string) string {
	// lightweight redaction without importing net/url errors on bad URLs
	if i := strings.Index(raw, "@"); i > 0 {
		schemeEnd := strings.Index(raw, "://")
		if schemeEnd >= 0 {
			return raw[:schemeEnd+3] + "***:***" + raw[i:]
		}
	}
	return raw
}

// Connect starts the MQTT client (idempotent).
func (s *Service) Connect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		return
	}
	opts := mqtt.NewClientOptions().
		AddBroker(s.env.MQTTURL).
		SetClientID(s.env.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectTimeout(30 * time.Second).
		SetKeepAlive(60 * time.Second)
	if s.env.MQTTUsername != "" {
		opts.SetUsername(s.env.MQTTUsername)
	}
	if s.env.MQTTPassword != "" {
		opts.SetPassword(s.env.MQTTPassword)
	}
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		s.mu.Lock()
		s.connected = true
		s.lastError = ""
		s.mu.Unlock()
		s.subscribeLocked(c)
		s.logger.Info("mqtt_connected")
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		s.mu.Lock()
		s.connected = false
		if err != nil {
			s.lastError = err.Error()
		}
		s.mu.Unlock()
	})
	opts.SetDefaultPublishHandler(s.messageHandler())

	cli := mqtt.NewClient(opts)
	token := cli.Connect()
	_ = token.WaitTimeout(35 * time.Second)
	if token.Error() != nil {
		s.lastError = token.Error().Error()
		s.logger.Error("mqtt_connect_error: ", token.Error())
	}
	s.client = cli
}

func (s *Service) subscribeLocked(c mqtt.Client) {
	s.subscribeMu.Lock()
	defer s.subscribeMu.Unlock()
	pat := s.env.FloraDevicesSubscribeTopic
	tok := c.Subscribe(pat, 1, nil)
	_ = tok.WaitTimeout(10 * time.Second)
	if tok.Error() != nil {
		s.logger.Error("mqtt_subscribe_error: ", tok.Error())
	}
}

func (s *Service) messageHandler() mqtt.MessageHandler {
	return func(_ mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		pattern := s.env.FloraDevicesSubscribeTopic
		segments := mqtttopic.ExtractWildcardSegments(pattern, topic)
		if segments == nil {
			return
		}
		logicalID := mqtttopic.CompositeDeviceID(segments)
		parsed := parseDevicePayload(msg.Payload())
		now := time.Now().UTC().Format(time.RFC3339Nano)
		var catalogRow string
		if dr, ok := mqtttopic.HiveIdentityFromSegments(segments); ok {
			rows, err := s.dev.ListByLogicalDeviceIDGlobally(dr)
			if err != nil {
				s.logger.Error("mqtt_device_lookup_error: ", err)
			} else if len(rows) == 0 {
				// unknown logical id — still keep a registry entry keyed by topic segment for ops
			} else if len(rows) > 1 {
				s.logger.Error("mqtt_device_lookup_ambiguous: device_id matches multiple rows")
			} else {
				catalogRow = rows[0].ID
			}
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		entry := &RegistryEntry{
			ID:         logicalID,
			LastSeenAt: now,
			LastTopic:  topic,
		}
		if catalogRow != "" {
			entry.DeviceRowID = catalogRow
		}
		if parsed.kind == "heartbeat" {
			entry.PresenceKind = PresenceTTL
			if parsed.meta != nil {
				entry.Meta = parsed.meta
			}
		} else {
			entry.PresenceKind = PresenceExplicit
			entry.ExplicitConnected = &parsed.connected
			if parsed.meta != nil {
				entry.Meta = parsed.meta
			}
		}
		s.registry[logicalID] = entry
	}
}

type parsedPayload struct {
	kind      string
	connected bool
	meta      interface{}
}

func parseDevicePayload(payload []byte) parsedPayload {
	s := strings.TrimSpace(string(payload))
	if s == "" {
		return parsedPayload{kind: "heartbeat"}
	}
	var j interface{}
	if err := json.Unmarshal([]byte(s), &j); err == nil {
		if m, ok := j.(map[string]interface{}); ok {
			if looksLikeFloraHeartbeat(m) {
				return parsedPayload{kind: "heartbeat", meta: m}
			}
			if v, ok := m["connected"].(bool); ok {
				return parsedPayload{kind: "status", connected: v, meta: m}
			}
			if v, ok := m["online"].(bool); ok {
				return parsedPayload{kind: "status", connected: v, meta: m}
			}
			if st, ok := m["state"].(string); ok {
				l := strings.ToLower(st)
				if l == "offline" || l == "disconnected" {
					return parsedPayload{kind: "status", connected: false, meta: m}
				}
				if l == "online" || l == "connected" {
					return parsedPayload{kind: "status", connected: true, meta: m}
				}
			}
			return parsedPayload{kind: "heartbeat", meta: m}
		}
	}
	lower := strings.ToLower(s)
	if lower == "offline" || lower == "false" || lower == "0" || lower == "disconnected" {
		return parsedPayload{kind: "status", connected: false}
	}
	if lower == "online" || lower == "true" || lower == "1" || lower == "connected" {
		return parsedPayload{kind: "status", connected: true}
	}
	return parsedPayload{kind: "heartbeat", meta: map[string]string{"raw": s}}
}

func looksLikeFloraHeartbeat(m map[string]interface{}) bool {
	if _, ok := m["ts"]; ok {
		return true
	}
	if _, ok := m["dht_status"]; ok {
		return true
	}
	if _, ok := m["registered_at"]; ok {
		return true
	}
	return false
}

func lastSeenWithinTTL(iso string, ttl time.Duration) bool {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		t, err = time.Parse(time.RFC3339, iso)
	}
	if err != nil {
		return false
	}
	return time.Since(t) <= ttl
}

func entryConnected(e *RegistryEntry, ttlSec int) bool {
	if e.PresenceKind == PresenceExplicit && e.ExplicitConnected != nil {
		return *e.ExplicitConnected
	}
	return lastSeenWithinTTL(e.LastSeenAt, time.Duration(ttlSec)*time.Second)
}

func toPublic(e *RegistryEntry, ttlSec int) PublicDevice {
	p := PublicDevice{
		ID:         e.ID,
		Connected:  entryConnected(e, ttlSec),
		LastSeenAt: e.LastSeenAt,
		LastTopic:  e.LastTopic,
	}
	if e.DeviceRowID != "" {
		p.Identity = &Identity{DeviceRowID: e.DeviceRowID}
	}
	if e.Meta != nil {
		p.Telemetry = e.Meta
	}
	return p
}

// ListLiveDevices returns registry entries filtered for the HTTP API.
func (s *Service) ListLiveDevices(includeOffline bool, allowedDeviceRowIDs []string) []PublicDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ttl := s.env.FloraDeviceHeartbeatTTLSec
	out := make([]PublicDevice, 0, len(s.registry))
	for _, e := range s.registry {
		p := toPublic(e, ttl)
		if !includeOffline && !p.Connected {
			continue
		}
		if allowedDeviceRowIDs != nil {
			if e.DeviceRowID == "" {
				continue
			}
			rowID := e.DeviceRowID
			allowed := false
			for _, a := range allowedDeviceRowIDs {
				if a == rowID {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// GetState returns MQTT connection state for the API.
func (s *Service) GetState() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var le *string
	if s.lastError != "" {
		le = &s.lastError
	}
	return State{
		Connected: s.connected,
		ClientID:  s.env.MQTTClientID,
		URL:       redactURL(s.env.MQTTURL),
		LastError: le,
	}
}

// Publish publishes a payload to a topic (normalized with prefix rules).
func (s *Service) Publish(topic string, payload interface{}, qos *int, retain bool) (string, int, bool, int, error) {
	s.mu.RLock()
	cli := s.client
	ok := s.connected
	s.mu.RUnlock()
	if cli == nil || !ok {
		return "", 0, false, 0, errNotConnected{}
	}
	norm, err := mqtttopic.NormalizePublishTopic(topic, s.env.FloraTopicPrefix)
	if err != nil {
		return "", 0, false, 0, err
	}
	q := s.env.MQTTDefaultQoS
	if qos != nil {
		q = *qos
	}
	if q < 0 || q > 2 {
		return "", 0, false, 0, errBadQoS{}
	}
	data, err := encodePayload(payload)
	if err != nil {
		return "", 0, false, 0, err
	}
	token := cli.Publish(norm, byte(q), retain, data)
	_ = token.WaitTimeout(10 * time.Second)
	if token.Error() != nil {
		return "", 0, false, 0, token.Error()
	}
	return norm, q, retain, len(data), nil
}

func encodePayload(payload interface{}) ([]byte, error) {
	if payload == nil {
		return []byte{}, nil
	}
	switch v := payload.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case json.RawMessage:
		return v, nil
	default:
		return json.Marshal(v)
	}
}

type errNotConnected struct{}

func (e errNotConnected) Error() string {
	return "MQTT not connected"
}

type errBadQoS struct{}

func (e errBadQoS) Error() string {
	return "qos must be 0, 1, or 2"
}

// Disconnect closes the client and clears registry.
func (s *Service) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return
	}
	s.client.Disconnect(250)
	s.client = nil
	s.connected = false
	s.registry = make(map[string]*RegistryEntry)
}
