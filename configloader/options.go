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

// WithNameTag sets the struct tag that can be used to override the environment variable name
// for a field. If a field has this tag, its value will be used as the exact environment variable
// name, ignoring any prefix and name conversion.
//
// Example:
//
//	type Config struct {
//	    Host string `env:"SERVICE_HOST"` // Will look for SERVICE_HOST environment variable
//	}
//	configloader.Load(&cfg, WithNameTag("env"))
func WithNameTag(tag string) Option {
	return func(l *Loader) {
		l.nameTag = tag
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
