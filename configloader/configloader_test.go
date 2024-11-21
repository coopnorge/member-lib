package configloader

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string `env:"DATABASE_USER"`
	Password string
}

type AppConfig struct {
	Environment string
	Debug       bool
	Database    DatabaseConfig
	AllowedIps  []string
	MaxRetries  int
	Timeout     float64
}

func TestLoad(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name    string
		envVars map[string]string
		want    AppConfig
		wantErr bool
	}{
		{
			name: "basic configuration",
			envVars: map[string]string{
				"ENVIRONMENT":       "production",
				"DEBUG":             "true",
				"DATABASE_HOST":     "localhost",
				"DATABASE_PORT":     "5432",
				"DATABASE_USER":     "admin",
				"DATABASE_PASSWORD": "secret",
				"ALLOWED_IPS":       "192.168.1.1,192.168.1.2",
				"MAX_RETRIES":       "3",
				"TIMEOUT":           "5.5",
			},
			want: AppConfig{
				Environment: "production",
				Debug:       true,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "admin",
					Password: "secret",
				},
				AllowedIps: []string{"192.168.1.1", "192.168.1.2"},
				MaxRetries: 3,
				Timeout:    5.5,
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				// Missing DATABASE_HOST
			},
			want:    AppConfig{},
			wantErr: true,
		},
		{
			name: "invalid type conversion",
			envVars: map[string]string{
				"ENVIRONMENT":   "production",
				"DATABASE_PORT": "not_a_number",
			},
			want:    AppConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment before each test
			os.Clearenv()

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Run loader
			got := &AppConfig{}
			err := Load(got, WithNameTag("env"))
			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Loader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Compare results
			if got.Environment != tt.want.Environment {
				t.Errorf("Environment = %v, want %v", got.Environment, tt.want.Environment)
			}
			if got.Debug != tt.want.Debug {
				t.Errorf("Debug = %v, want %v", got.Debug, tt.want.Debug)
			}
			if got.Database.Host != tt.want.Database.Host {
				t.Errorf("Database.Host = %v, want %v", got.Database.Host, tt.want.Database.Host)
			}
			if got.Database.Port != tt.want.Database.Port {
				t.Errorf("Database.Port = %v, want %v", got.Database.Port, tt.want.Database.Port)
			}
			if got.Database.Username != tt.want.Database.Username {
				t.Errorf("Database.Username = %v, want %v", got.Database.Username, tt.want.Database.Username)
			}
			if got.Database.Password != tt.want.Database.Password {
				t.Errorf("Database.Password = %v, want %v", got.Database.Password, tt.want.Database.Password)
			}
			if len(got.AllowedIps) != len(tt.want.AllowedIps) {
				t.Errorf("AllowedIPs len = %v, want %v", len(got.AllowedIps), len(tt.want.AllowedIps))
			} else {
				for i := range got.AllowedIps {
					if got.AllowedIps[i] != tt.want.AllowedIps[i] {
						t.Errorf("AllowedIPs[%d] = %v, want %v", i, got.AllowedIps[i], tt.want.AllowedIps[i])
					}
				}
			}
			if got.MaxRetries != tt.want.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", got.MaxRetries, tt.want.MaxRetries)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %v, want %v", got.Timeout, tt.want.Timeout)
			}
		})
	}
}

func TestFails_OnConfigNotBeingStruct(t *testing.T) {
	conf := AppConfig{}
	assert.Error(t, Load(conf), "config should fail as it's not a pointer")
}

type JsonExm struct {
	Name  string `json:"name"`
	Split string `json:"split,omitempty"`
}

func TestNameJson(t *testing.T) {
	t.Setenv("name", "hehe")
	t.Setenv("split", "splitval")
	var conf JsonExm
	err := Load(&conf, WithNameTag("json"))
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "hehe", conf.Name, "name is not the same as hehe")
	assert.Equal(t, "splitval", conf.Split, "split is incorrenclty read")
}

type ComplexConfig struct {
	Env     string `mapstructure:"dd_env" json:"dd_env,omitempty"`
	Service string `mapstructure:"dd_service" json:"dd_service,omitempty"`
	// Field Name is not the same as mapstructure tag val
	ServiceVersion       string `mapstructure:"dd_version" json:"dd_service_version,omitempty"`
	DSD                  string `mapstructure:"dd_dogstatsd_url" json:"dd_dsd,omitempty"`
	APM                  string `mapstructure:"dd_trace_agent_url" json:"dd_apm,omitempty"`
	EnableExtraProfiling bool   `mapstructure:"dd_enable_extra_profiling" json:"dd_enable_extra_profiling,omitempty"`
}

func CustomGetenv(val string) string {
	return os.Getenv(strings.ToUpper(val))
}

func TestComplexLoading(t *testing.T) {
	t.Setenv("DD_ENV", "prod")
	t.Setenv("DD_VERSION", "1.0.0")
	t.Setenv("DD_DOGSTATSD_URL", "hehe://someproto")
	t.Setenv("DD_ENABLE_EXTRA_PROFILING", "TRUE")

	var conf ComplexConfig

	err := Load(&conf, WithNameTag("mapstructure"), WithEnv(CustomGetenv))
	assert.NoError(t, err, "Loading should not fail")
	assert.Equal(t, "prod", conf.Env, "Env should be prod")
	assert.Equal(t, "1.0.0", conf.ServiceVersion)
	assert.Equal(t, "", conf.Service, "Service should be zero value")
	assert.Equal(t, "hehe://someproto", conf.DSD, "Dsd was not picked up ")
}

type ConfigWP struct {
	Pel *PointerElement
}

type PointerElement struct {
	Name string
}

func TestLoadingWithPointerNested(t *testing.T) {
	var conf ConfigWP
	err := Load(&conf)

	assert.NoError(t, err, "Expected error free Load")
}
