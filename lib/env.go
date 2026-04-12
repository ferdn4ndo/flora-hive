package lib

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Env holds application configuration (Flora Hive + shared template fields).
type Env struct {
	ServerPort  string `mapstructure:"PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	SentryDsn   string `mapstructure:"SENTRY_DSN"`

	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASS"`
	PostgresSSLMode  string `mapstructure:"POSTGRES_SSLMODE"`

	MQTTURL        string `mapstructure:"MQTT_URL"`
	MQTTUsername   string `mapstructure:"MQTT_USERNAME"`
	MQTTPassword   string `mapstructure:"MQTT_PASSWORD"`
	MQTTClientID   string `mapstructure:"MQTT_CLIENT_ID"`
	MQTTDefaultQoS int    `mapstructure:"MQTT_DEFAULT_QOS"`

	FloraTopicPrefix           string `mapstructure:"FLORA_TOPIC_PREFIX"`
	FloraDevicesSubscribeTopic string `mapstructure:"FLORA_DEVICES_SUBSCRIBE_TOPIC"`
	FloraDeviceHeartbeatTTLSec int    `mapstructure:"FLORA_DEVICE_HEARTBEAT_TTL_SEC"`

	UserverAuthHost        string `mapstructure:"USERVER_AUTH_HOST"`
	UserverAuthSystemName  string `mapstructure:"USERVER_AUTH_SYSTEM_NAME"`
	UserverAuthSystemToken string `mapstructure:"USERVER_AUTH_SYSTEM_TOKEN"`

	HiveAPIKeysRaw string `mapstructure:"HIVE_API_KEYS"`
}

// ParseHiveAPIKeysFromRaw splits HIVE_API_KEYS (comma-separated) into trimmed, non-empty keys.
func ParseHiveAPIKeysFromRaw(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// HiveAPIKeys returns trimmed, non-empty API keys from HIVE_API_KEYS (comma-separated).
func (e *Env) HiveAPIKeys() []string {
	return ParseHiveAPIKeysFromRaw(e.HiveAPIKeysRaw)
}

// UserverConfigured is true when uServer-Auth integration env vars are set.
func (e *Env) UserverConfigured() bool {
	h := strings.TrimSuffix(strings.TrimSpace(e.UserverAuthHost), "/")
	return h != "" &&
		strings.TrimSpace(e.UserverAuthSystemName) != "" &&
		strings.TrimSpace(e.UserverAuthSystemToken) != ""
}

// UserverAuthBase returns the trimmed auth host without trailing slash.
func (e *Env) UserverAuthBase() string {
	return strings.TrimSuffix(strings.TrimSpace(e.UserverAuthHost), "/")
}

// IsLocal mirrors the template: local environment disables some production-only behavior.
func (e *Env) IsLocal() bool {
	return e.Environment == "local"
}

// NewEnv loads configuration from the environment.
func NewEnv() Env {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "local")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("MQTT_DEFAULT_QOS", 1)
	viper.SetDefault("MQTT_CLIENT_ID", "flora-hive")
	viper.SetDefault("FLORA_TOPIC_PREFIX", "flora")
	viper.SetDefault("FLORA_DEVICE_HEARTBEAT_TTL_SEC", 180)
	viper.AutomaticEnv()
	// Unmarshal only merges keys that appear in AllKeys() (defaults, file, BindEnv).
	// AutomaticEnv alone does not register unknown keys, so bind env vars we read from the OS.
	for _, k := range []string{
		"PORT", "ENVIRONMENT", "LOG_LEVEL", "SENTRY_DSN",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB", "POSTGRES_USER", "POSTGRES_SSLMODE",
		"MQTT_URL", "MQTT_USERNAME", "MQTT_PASSWORD", "MQTT_CLIENT_ID", "MQTT_DEFAULT_QOS",
		"FLORA_TOPIC_PREFIX", "FLORA_DEVICES_SUBSCRIBE_TOPIC", "FLORA_DEVICE_HEARTBEAT_TTL_SEC",
		"USERVER_AUTH_HOST", "USERVER_AUTH_SYSTEM_NAME", "USERVER_AUTH_SYSTEM_TOKEN",
		"HIVE_API_KEYS",
	} {
		if err := viper.BindEnv(k); err != nil {
			log.Fatal("bind env: ", err)
		}
	}
	// Password: accept POSTGRES_PASS or POSTGRES_PASSWORD (same as manual merge below).
	if err := viper.BindEnv("POSTGRES_PASS", "POSTGRES_PASS", "POSTGRES_PASSWORD"); err != nil {
		log.Fatal("bind env: ", err)
	}
	if strings.TrimSpace(viper.GetString("POSTGRES_PASS")) == "" && os.Getenv("POSTGRES_PASSWORD") != "" {
		viper.Set("POSTGRES_PASS", strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")))
	}

	var env Env
	if err := viper.Unmarshal(&env); err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	if env.PostgresSSLMode == "" {
		env.PostgresSSLMode = "disable"
	}
	if env.FloraTopicPrefix != "" {
		env.FloraTopicPrefix = strings.TrimSuffix(strings.TrimSpace(env.FloraTopicPrefix), "/")
	}
	if env.FloraDevicesSubscribeTopic == "" {
		p := strings.TrimSpace(env.FloraTopicPrefix)
		if p == "" {
			env.FloraDevicesSubscribeTopic = "+/heartbeat"
		} else {
			env.FloraDevicesSubscribeTopic = fmt.Sprintf("%s/+/heartbeat", p)
		}
	} else {
		env.FloraDevicesSubscribeTopic = strings.TrimSpace(env.FloraDevicesSubscribeTopic)
	}

	mustHave := []struct {
		name, val string
	}{
		{"POSTGRES_HOST", env.PostgresHost},
		{"POSTGRES_DB", env.PostgresDB},
		{"POSTGRES_USER", env.PostgresUser},
		{"MQTT_URL", env.MQTTURL},
	}
	for _, m := range mustHave {
		if strings.TrimSpace(m.val) == "" {
			log.Fatal("required environment variable missing: ", m.name)
		}
	}
	if env.MQTTDefaultQoS < 0 || env.MQTTDefaultQoS > 2 {
		log.Fatal("MQTT_DEFAULT_QOS must be 0, 1, or 2")
	}
	if env.FloraDeviceHeartbeatTTLSec < 10 || env.FloraDeviceHeartbeatTTLSec > 86400 {
		log.Fatal("FLORA_DEVICE_HEARTBEAT_TTL_SEC must be between 10 and 86400")
	}

	// Normalize port strings if viper gave numbers via env
	if _, err := strconv.Atoi(env.PostgresPort); err != nil {
		log.Fatal("invalid POSTGRES_PORT: ", env.PostgresPort)
	}
	if _, err := strconv.Atoi(env.ServerPort); err != nil {
		log.Fatal("invalid PORT: ", env.ServerPort)
	}

	return env
}
