package configloader

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	t.Run("support fields", func(t *testing.T) {
		var cfg struct {
			Foo string
		}
		require.NoError(t, os.Setenv("FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Foo)
	})

	t.Run("support field pointers", func(t *testing.T) {
		var cfg struct {
			Foo *string
		}
		require.NoError(t, os.Setenv("FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", *cfg.Foo)
	})

	t.Run("support slices", func(t *testing.T) {
		var cfg struct {
			Foo []string
		}
		require.NoError(t, os.Setenv("FOO", "foo,bar"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar"}, cfg.Foo)
	})

	t.Run("support inner struct", func(t *testing.T) {
		var cfg struct {
			Inner struct {
				Foo string
			}
		}
		require.NoError(t, os.Setenv("INNER_FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Inner.Foo)
	})

	t.Run("support inner struct pointers", func(t *testing.T) {
		var cfg struct {
			Inner *struct {
				Foo string
			}
		}
		require.NoError(t, os.Setenv("INNER_FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Inner.Foo)
	})

	t.Run("support embedded structs", func(t *testing.T) {
		type Test struct {
			Foo string
		}
		var cfg struct {
			Test
			Inner Test
		}
		require.NoError(t, os.Setenv("FOO", "foo"))
		require.NoError(t, os.Setenv("INNER_FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Inner.Foo)
		assert.Equal(t, "foo", cfg.Test)
		assert.Equal(t, "foo", cfg.Foo)
	})

	t.Run("fails if not pointer", func(t *testing.T) {
		var cfg struct {
			Foo string
		}
		assert.EqualError(t, Load(cfg), "val must be a pointer, got 'struct { Foo string }'")
	})

	t.Run("fails if nil pointer", func(t *testing.T) {
		type Test struct {
			Foo string
		}
		var cfg *Test
		assert.EqualError(t, Load(cfg), "val cannot be nil")
	})
}

func TestLoader_Load_structs(t *testing.T) {
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

func CustomGetenv(val string) (string, bool) {
	return os.LookupEnv(strings.ToUpper(val))
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

func TestWithPrefix(t *testing.T) {
	t.Run("supports prefix", func(t *testing.T) {
		var cfg struct {
			Val    string
			Struct struct {
				Val string
			}
		}
		require.NoError(t, os.Setenv("PREFIX_VAL", "bar"))
		require.NoError(t, os.Setenv("PREFIX_STRUCT_VAL", "bar"))
		err := Load(&cfg, WithPrefix("PREFIX"))
		assert.NoError(t, err)

		assert.Equal(t, "bar", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})
}

func TestWithNameTag(t *testing.T) {
	t.Run("uses value from name tag", func(t *testing.T) {
		var cfg struct {
			Val    string `env:"FOO"`
			Struct struct {
				Val string `env:"BAR"`
			}
		}
		require.NoError(t, os.Setenv("FOO", "foo"))
		require.NoError(t, os.Setenv("BAR", "bar"))
		err := Load(&cfg, WithNameTag("env"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})

	t.Run("support split tags", func(t *testing.T) {
		var cfg struct {
			Val string `env:"FOO,bar"`
		}
		require.NoError(t, os.Setenv("FOO", "foo"))
		defer os.Clearenv()

		err := Load(&cfg, WithNameTag("env"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
	})

	t.Run("uses value from name tag and prefix", func(t *testing.T) {
		var cfg struct {
			Val    string `env:"FOO"`
			Struct struct {
				Val string `env:"BAR"`
			}
		}
		require.NoError(t, os.Setenv("PREFIX_FOO", "foo"))
		require.NoError(t, os.Setenv("PREFIX_STRUCT_BAR", "bar"))
		defer os.Clearenv()

		err := Load(&cfg, WithNameTag("env"), WithPrefix("PREFIX"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})

	t.Run("uses default value", func(t *testing.T) {
		var cfg struct {
			Val        string `env:"FOO"`
			ValDefault string `env:"FOO_DEF" default:"defaultFoo"`
			Struct     struct {
				Val        string `env:"BAR"`
				ValDefault string `env:"FOO_DEF" default:"defaultBar"`
			}
		}
		require.NoError(t, os.Setenv("PREFIX_FOO", "foo"))
		require.NoError(t, os.Setenv("PREFIX_BAR", "bar"))
		err := Load(&cfg,
			WithNameTag("env"),
			WithPrefix("PREFIX"),
			WithDefaultTag("default"),
		)
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
		assert.Equal(t, "defaultFoo", cfg.ValDefault)
		assert.Equal(t, "defaultBar", cfg.Struct.ValDefault)
	})
}

func TestWithDefaultTag(t *testing.T) {
	t.Run("uses value from default", func(t *testing.T) {
		var cfg struct {
			Foo string `default:"foo"`
			Bar struct {
				Baz string `default:"baz"`
			}
		}

		err := Load(&cfg, WithEnv(CustomGetenv))
		assert.NoError(t, err)
		assert.Equal(t, "foo", cfg.Foo)
		assert.Equal(t, "baz", cfg.Bar.Baz)
	})

	t.Run("uses custom tag", func(t *testing.T) {
		var cfg struct {
			Foo string `default2:"foo"`
			Bar struct {
				Baz string `default2:"baz"`
			}
		}

		err := Load(&cfg, WithDefaultTag("default2"), WithEnv(CustomGetenv))
		assert.NoError(t, err)
		assert.Equal(t, "foo", cfg.Foo)
		assert.Equal(t, "baz", cfg.Bar.Baz)
	})

	t.Run("uses value from env", func(t *testing.T) {
		defer os.Clearenv()

		var cfg struct {
			Foo string `default:"foo"`
			Bar struct {
				Baz string `default:"baz"`
			}
		}

		require.NoError(t, os.Setenv("FOO", "fooEnv"))
		require.NoError(t, os.Setenv("BAR_BAZ", "bazEnv"))

		err := Load(&cfg, WithEnv(CustomGetenv))
		assert.NoError(t, err)
		assert.Equal(t, "fooEnv", cfg.Foo)
		assert.Equal(t, "bazEnv", cfg.Bar.Baz)
	})

	t.Run("works with custom types", func(t *testing.T) {
		var cfg struct {
			Foo *url.URL `default:"https://example.com/foo"`
			Bar struct {
				Baz *url.URL `default:"https://example.com/baz"`
			}
		}

		err := Load(&cfg, WithEnv(CustomGetenv))
		assert.NoError(t, err)

		foo, err := url.Parse("https://example.com/foo")
		require.NoError(t, err)
		baz, err := url.Parse("https://example.com/baz")
		require.NoError(t, err)

		assert.Equal(t, foo, cfg.Foo)
		assert.Equal(t, baz, cfg.Bar.Baz)
	})
}

func TestWithTypeHandler(t *testing.T) {
	t.Run("fails with unsupported types", func(t *testing.T) {
		type CustomType string
		var cfg struct {
			Val CustomType
		}
		require.NoError(t, os.Setenv("VAL", "foo"))
		defer os.Clearenv()
		err := Load(&cfg)
		assert.EqualError(t, err,
			`failed to load struct { Val configloader.CustomType }:
error processing field Val (configloader.CustomType): unsupported type configloader.CustomType`)
	})

	t.Run("fails with error from handler", func(t *testing.T) {
		type CustomType string
		var cfg struct {
			Val CustomType
		}
		require.NoError(t, os.Setenv("VAL", "foo"))
		defer os.Clearenv()

		expectedErr := errors.New("expected")
		err := Load(&cfg, WithTypeHandler(func(val string) (CustomType, error) {
			assert.Equal(t, "foo", val)
			return "", expectedErr
		}))
		assert.ErrorIs(t, err, expectedErr)
		assert.EqualError(t, err, `failed to load struct { Val configloader.CustomType }:
error processing field Val (configloader.CustomType): expected`)
	})

	t.Run("support custom handler for primitive", func(t *testing.T) {
		type AnotherString string
		var cfg struct {
			Val      AnotherString
			ValPtr   *AnotherString
			ValSlice []AnotherString
		}
		require.NoError(t, os.Setenv("VAL", "foo"))
		require.NoError(t, os.Setenv("VAL_PTR", "foo"))
		require.NoError(t, os.Setenv("VAL_SLICE", "foo,foo"))
		defer os.Clearenv()

		err := Load(&cfg, WithTypeHandler(func(val string) (AnotherString, error) {
			assert.Equal(t, "foo", val)
			return "notFoo", nil
		}))
		assert.NoError(t, err)

		assert.Equal(t, AnotherString("notFoo"), cfg.Val)
		assert.Equal(t, AnotherString("notFoo"), *cfg.ValPtr)
		assert.Equal(t, []AnotherString{"notFoo", "notFoo"}, cfg.ValSlice)
	})

	t.Run("support custom handler for structs", func(t *testing.T) {
		type Test struct {
			Val string
		}
		var cfg struct {
			Val      Test
			ValPtr   *Test
			ValSlice []Test
		}
		require.NoError(t, os.Setenv("VAL", "foo"))
		require.NoError(t, os.Setenv("VAL_PTR", "foo"))
		require.NoError(t, os.Setenv("VAL_SLICE", "foo,foo"))
		defer os.Clearenv()

		err := Load(&cfg, WithTypeHandler(func(val string) (Test, error) {
			assert.Equal(t, "foo", val)
			return Test{Val: "notFoo"}, nil
		}))
		assert.NoError(t, err)

		assert.Equal(t, Test{Val: "notFoo"}, cfg.Val)
		assert.Equal(t, Test{Val: "notFoo"}, *cfg.ValPtr)
		assert.Equal(t, []Test{{Val: "notFoo"}, {Val: "notFoo"}}, cfg.ValSlice)
	})
}

func TestWithEnv(t *testing.T) {
	t.Run("support custom env", func(t *testing.T) {
		var cfg struct {
			Val string
		}

		err := Load(&cfg, WithEnv(func(s string) (string, bool) {
			assert.Equal(t, "VAL", s)
			return "notFoo", true
		}))
		require.NoError(t, err)
		assert.Equal(t, "notFoo", cfg.Val)
	})
}
