package configloader

import (
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
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
			os.Clearenv()
			for k, v := range tt.envVars {
				require.NoError(t, os.Setenv(k, v))
			}

			got := &AppConfig{}
			err := Load(got, WithNameTag("env"))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Environment, got.Environment)
			assert.Equal(t, tt.want.Debug, got.Debug)
			assert.Equal(t, tt.want.Database.Host, got.Database.Host)
			assert.Equal(t, tt.want.Database.Port, got.Database.Port)
			assert.Equal(t, tt.want.Database.Username, got.Database.Username)
			assert.Equal(t, tt.want.Database.Password, got.Database.Password)
			assert.Equal(t, tt.want.AllowedIps, got.AllowedIps)
			assert.Equal(t, tt.want.MaxRetries, got.MaxRetries)
			assert.Equal(t, tt.want.Timeout, got.Timeout)
		})
	}
}

func TestFails_OnConfigNotBeingStruct(t *testing.T) {
	type AppConfig struct {
		Name string
	}
	conf := AppConfig{}
	assert.Error(t, Load(conf), "config should fail as it's not a pointer")
}

func TestNameJson(t *testing.T) {
	type JsonExm struct {
		Name  string `json:"name"`
		Split string `json:"split,omitempty"`
	}

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
	// Field Name is not the same as mapstructure nameTag val
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

func TestWithPrefixTag(t *testing.T) {
	var cfg struct {
		Foo string
		Bar struct {
			Baz string
		}
	}

	require.NoError(t, os.Setenv("TEST_FOO", "fooEnv"))
	require.NoError(t, os.Setenv("TEST_BAR_BAZ", "bazEnv"))

	err := Load(&cfg, WithPrefix("TEST"), WithEnv(CustomGetenv))
	assert.NoError(t, err)
	assert.Equal(t, "fooEnv", cfg.Foo)
	assert.Equal(t, "bazEnv", cfg.Bar.Baz)

}

func TestWithDefaultTag(t *testing.T) {
	t.Run("uses value from default", func(t *testing.T) {
		var cfg struct {
			Foo string `default:"foo"`
			Bar struct {
				Baz string `default:"baz"`
			}
		}

		err := Load(&cfg, WithDefaultTag("default"), WithEnv(CustomGetenv))
		assert.NoError(t, err)
		assert.Equal(t, "foo", cfg.Foo)
		assert.Equal(t, "baz", cfg.Bar.Baz)
	})

	t.Run("uses value from env", func(t *testing.T) {
		var cfg struct {
			Foo string `default:"foo"`
			Bar struct {
				Baz string `default:"baz"`
			}
		}

		require.NoError(t, os.Setenv("FOO", "fooEnv"))
		require.NoError(t, os.Setenv("BAZ", "bazEnv"))

		err := Load(&cfg, WithDefaultTag("default"), WithEnv(CustomGetenv))
		assert.NoError(t, err)
		assert.Equal(t, "fooEnv", cfg.Foo)
		assert.Equal(t, "bazEnv", cfg.Bar.Baz)
	})
}

func TestLoadingWithPointerNested(t *testing.T) {
	type SecondElement struct {
		Url string
	}

	type PointerElement struct {
		Name string
		Sec  *SecondElement
	}

	type ConfigWP struct {
		Pel *PointerElement
	}

	// NOTE: This is how we use the lib.
	t.Setenv("PEL_NAME", "alfredo")
	t.Setenv("PEL_SEC_URL", "www.pelsec.url")
	var conf ConfigWP
	err := Load(&conf)

	assert.NoError(t, err, "Expected error free Load")
	assert.Equal(t, "alfredo", conf.Pel.Name)
	assert.Equal(t, "www.pelsec.url", conf.Pel.Sec.Url)
}

func TestLoadingComplexType(t *testing.T) {
	// Complex types
	type ComplexTypeEx struct {
		Timeout time.Duration
		URL     url.URL
		IPAddr  net.IP
	}

	t.Setenv("TIMEOUT", "1s")
	t.Setenv("URL_HOST", "localhost:8080")
	t.Setenv("IP_ADDR", "192.168.1.1")

	var conf ComplexTypeEx
	err := Load(&conf,
		WithTypeHandler[time.Duration](func(s string) (time.Duration, error) {
			return time.ParseDuration(s)
		}), // IP address handler
		WithTypeHandler[net.IP](func(s string) (net.IP, error) {
			return net.ParseIP(s), nil
		}))

	assert.NoError(t, err, "Expected error free Load")
	sec, _ := time.ParseDuration("1s")
	assert.Equal(t, sec, conf.Timeout, "Timeout was not read properly")
	assert.Equal(t, "localhost:8080", conf.URL.Host)
	assert.Equal(t, "192.168.1.1", conf.IPAddr.String())
}
