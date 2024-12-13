package configloader

import "reflect"

type Option func(*Loader)

// WithTypeHandler registers a custom type conversion function for a specific type T.
// The function should convert a string environment value to type T, returning an error
// if the conversion fails.
//
// Example:
//
//	configloader.Load(&cfg, WithTypeHandler(func(s string) (time.Duration, error) {
//	    return time.ParseDuration(s)
//	}))
func WithTypeHandler[T any](f func(string) (T, error)) Option {
	return func(l *Loader) {
		var zero T
		l.handlers[reflect.TypeOf(zero)] = func(value string) (any, error) {
			return f(value)
		}
	}
}

// WithNameTag sets the struct tag used to override a field's name in the environment variable path.
// The tag value replaces just the field's name segment while still following the normal path construction
// rules (prefix + path + name).
//
// Example with WithNameTag("name"):
//
//	type Config struct {
//	    Database struct {
//	        Host string `name:"HOSTNAME"` // Looks for DATABASE_HOSTNAME
//	    }
//	}
//	configloader.Load(&cfg)
//
// Example with both prefix and name tag:
//
//	type Config struct {
//	    Database struct {
//	        Host string `name:"HOSTNAME"` // Looks for APP_DATABASE_HOSTNAME
//	    }
//	}
//	configloader.Load(&cfg, WithPrefix("APP"))
func WithNameTag(tag string) Option {
	return func(l *Loader) {
		l.nameTag = tag
	}
}

// WithEnvTag sets the struct tag used to completely override the environment variable name for a field.
// When a field has this tag, its value is used as-is for the environment variable name, bypassing all
// other name construction rules including prefixes and path building.
//
// Example with WithEnvTag("env"):
//
//	type Config struct {
//	    Database struct {
//	        // Despite nesting, looks directly for "DB_HOST"
//	        Host string `env:"DB_HOST"`
//	    }
//	}
//	configloader.Load(&cfg)
//
// Example showing prefix is ignored with env tag:
//
//	type Config struct {
//	    Database struct {
//	        // Still only looks for "DB_HOST", prefix is not applied
//	        Host string `env:"DB_HOST"`
//	    }
//	}
//	configloader.Load(&cfg, WithPrefix("APP"))
func WithEnvTag(tag string) Option {
	return func(l *Loader) {
		l.envTag = tag
	}
}

// WithDefaultTag sets the struct tag used for specifying default values.
// If an environment variable is not found, the value of this tag will be used instead.
//
// Example:
//
//	type Config struct {
//	    Port int `default:"8080"`  // Will use 8080 if PORT is not set
//	}
//	configloader.Load(&cfg, WithDefaultTag("default"))
func WithDefaultTag(tag string) Option {
	return func(l *Loader) {
		l.defaultTag = tag
	}
}

// WithPrefix sets a prefix that will be prepended to all environment variable names.
// The prefix and field name will be joined with an underscore.
//
// Example:
//
//	type Config struct {
//	    Port int  // Will look for "APP_PORT" environment variable
//	}
//	configloader.Load(&cfg, WithPrefix("APP"))
func WithPrefix(prefix string) Option {
	return func(l *Loader) {
		l.prefix = prefix
	}
}

// WithEnv sets a custom function for looking up environment variables.
// This is primarily useful for testing or when environment variables need to be
// sourced from somewhere other than os.LookupEnv.
//
// The function should return the value and a boolean indicating whether the variable
// was found, similar to os.LookupEnv.
func WithEnv(env func(string) (string, bool)) Option {
	return func(l *Loader) {
		l.env = env
	}
}
