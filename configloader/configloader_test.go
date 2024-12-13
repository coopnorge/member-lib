package configloader

import (
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	t.Run("support fields", func(t *testing.T) {
		var cfg struct {
			Foo string
		}
		t.Setenv("FOO", "foo")
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Foo)
	})

	t.Run("support field pointers", func(t *testing.T) {
		var cfg struct {
			Foo *string
		}
		t.Setenv("FOO", "foo")
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", *cfg.Foo)
	})

	t.Run("support slices", func(t *testing.T) {
		var cfg struct {
			Foo []string
		}
		t.Setenv("FOO", "foo,bar")
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
		t.Setenv("INNER_FOO", "foo")
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
		t.Setenv("INNER_FOO", "foo")
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
		t.Setenv("FOO", "foo")
		t.Setenv("INNER_FOO", "foo")
		defer os.Clearenv()

		err := Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "foo", cfg.Inner.Foo)
		assert.Equal(t, "foo", cfg.Test.Foo)
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

	t.Run("fails if environment not found", func(t *testing.T) {
		type Test struct {
			Foo string
			Bar string
		}
		var cfg Test
		assert.EqualError(t, Load(&cfg), `failed to load configloader.Test:
error processing field Foo (string): environment variable FOO not found
error processing field Bar (string): environment variable BAR not found`)
	})
}

func TestWithPrefixTag(t *testing.T) {
	var cfg struct {
		Foo string
		Bar struct {
			Baz string
		}
	}

	t.Setenv("TEST_FOO", "fooEnv")
	t.Setenv("TEST_BAR_BAZ", "bazEnv")

	err := Load(&cfg, WithPrefix("TEST"))
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
		t.Setenv("PREFIX_VAL", "bar")
		t.Setenv("PREFIX_STRUCT_VAL", "bar")
		err := Load(&cfg, WithPrefix("PREFIX"))
		assert.NoError(t, err)

		assert.Equal(t, "bar", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})
}

func TestWithNameTag(t *testing.T) {
	t.Run("uses value from name tag", func(t *testing.T) {
		var cfg struct {
			Val    string `name:"FOO"`
			Struct struct {
				Val string `name:"BAR"`
			}
		}
		t.Setenv("FOO", "foo")
		t.Setenv("STRUCT_BAR", "bar")
		err := Load(&cfg, WithNameTag("name"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})

	t.Run("uses value from name tag and prefix", func(t *testing.T) {
		var cfg struct {
			Val    string `name:"FOO"`
			Struct struct {
				Val string `name:"BAR"`
			}
		}
		t.Setenv("PREFIX_FOO", "foo")
		t.Setenv("PREFIX_STRUCT_BAR", "bar")
		defer os.Clearenv()

		err := Load(&cfg, WithNameTag("name"), WithPrefix("PREFIX"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})

	t.Run("uses default value", func(t *testing.T) {
		var cfg struct {
			Val        string `name:"FOO"`
			ValDefault string `name:"FOO_DEF" default:"defaultFoo"`
			Struct     struct {
				Val        string `name:"BAR"`
				ValDefault string `name:"FOO_DEF" default:"defaultBar"`
			}
		}
		t.Setenv("FOO", "foo")
		t.Setenv("STRUCT_BAR", "bar")
		err := Load(&cfg,
			WithNameTag("name"),
			WithDefaultTag("default"),
		)
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
		assert.Equal(t, "defaultFoo", cfg.ValDefault)
		assert.Equal(t, "defaultBar", cfg.Struct.ValDefault)
	})
}

func TestWithEnvTag(t *testing.T) {
	t.Run("uses value from env tag", func(t *testing.T) {
		var cfg struct {
			Val    string `env:"FOO"`
			Struct struct {
				Val string `env:"BAR"`
			}
		}
		t.Setenv("FOO", "foo")
		t.Setenv("BAR", "bar")
		err := Load(&cfg, WithEnvTag("env"))
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
	})

	t.Run("uses value from name tag not the prefix", func(t *testing.T) {
		var cfg struct {
			Val    string `env:"FOO"`
			Struct struct {
				Val string `env:"BAR"`
			}
		}
		t.Setenv("FOO", "foo")
		t.Setenv("BAR", "bar")
		defer os.Clearenv()

		err := Load(&cfg, WithEnvTag("env"), WithPrefix("PREFIX"))
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
		t.Setenv("FOO", "foo")
		t.Setenv("BAR", "bar")
		err := Load(&cfg,
			WithEnvTag("env"),
			WithDefaultTag("default"),
		)
		assert.NoError(t, err)

		assert.Equal(t, "foo", cfg.Val)
		assert.Equal(t, "bar", cfg.Struct.Val)
		assert.Equal(t, "defaultFoo", cfg.ValDefault)
		assert.Equal(t, "defaultBar", cfg.Struct.ValDefault)
	})

	t.Run("fails if using on struct", func(t *testing.T) {
		type Test struct {
			Val string
		}
		var cfg struct {
			Val    string `env:"FOO"`
			Struct Test   `env:"SHOULD_FAIL"`
		}
		t.Setenv("FOO", "foo")
		t.Setenv("BAR", "bar")
		err := Load(&cfg,
			WithEnvTag("env"),
			WithDefaultTag("default"),
		)
		assert.EqualError(t, err, "failed to load struct { Val string \"env:\\\"FOO\\\"\"; Struct configloader.Test \"env:\\\"SHOULD_FAIL\\\"\" }:\nerror processing field Struct.Val (string): ´env´ tag need to be at the end: [{Struct  configloader.Test env:\"SHOULD_FAIL\" 16 [1] false}]")
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

		err := Load(&cfg)
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

		err := Load(&cfg, WithDefaultTag("default2"))
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

		t.Setenv("FOO", "fooEnv")
		t.Setenv("BAR_BAZ", "bazEnv")

		err := Load(&cfg)
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

		err := Load(&cfg)
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
		t.Setenv("VAL", "foo")
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
		t.Setenv("VAL", "foo")
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

	t.Run("fails with element of slice", func(t *testing.T) {
		type CustomType string
		var cfg struct {
			Val []CustomType
		}
		t.Setenv("VAL", "foo,foo")
		defer os.Clearenv()

		expectedErr := errors.New("expected")
		err := Load(&cfg, WithTypeHandler(func(val string) (CustomType, error) {
			assert.Equal(t, "foo", val)
			return "", expectedErr
		}))
		assert.ErrorIs(t, err, expectedErr)
		assert.EqualError(t, err, "failed to load struct { Val []configloader.CustomType }:\nerror processing field Val ([]configloader.CustomType): expected\nexpected")
	})

	t.Run("support custom handler for primitive", func(t *testing.T) {
		type AnotherString string
		var cfg struct {
			Val      AnotherString
			ValPtr   *AnotherString
			ValSlice []AnotherString
		}
		t.Setenv("VAL", "foo")
		t.Setenv("VAL_PTR", "foo")
		t.Setenv("VAL_SLICE", "foo,foo")
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
		t.Setenv("VAL", "foo")
		t.Setenv("VAL_PTR", "foo")
		t.Setenv("VAL_SLICE", "foo,foo")
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
