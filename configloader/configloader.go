package configloader

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

type Loader struct {
	prefix, nameTag, defaultTag string

	// Type of conversion the struct field will take. Default is SnakeCase.
	fieldConversion func(string) string

	// Custom GetEnv function.
	env func(string) (string, bool)

	// Map of type handlers where key is reflect.Type and value is the handler function
	handlers map[reflect.Type]func(string) (any, error)
}

type Field struct {
	Value reflect.Value
	Path  []reflect.StructField
}

func (f *Field) Names() (names []string) {
	for _, field := range f.Path {
		names = append(names, field.Name)
	}
	return names
}

func (f *Field) String() string {
	return fmt.Sprintf("%s (%s)", strings.Join(f.Names(), "."), f.Value.Type().String())
}

type UnsupportedTypeError struct {
	reflect.Type
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type %s", e.Type)
}

func defaultLoader() *Loader {
	l := &Loader{
		handlers:        make(map[reflect.Type]func(string) (any, error), 0),
		fieldConversion: strcase.ToScreamingSnake,
		env:             os.LookupEnv,
		defaultTag:      "default",
	}
	for _, opt := range defaultTypeHandlers {
		opt(l)
	}
	return l
}

type Option func(*Loader)

var _ error = MissingEnvErr{}

type MissingEnvErr struct {
	Key   string
	Field Field
}

func (m MissingEnvErr) Error() string {
	return fmt.Sprintf("missing variable '%s' for the field '%s'", m.Key, m.Field.String())
}

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

// Load populates a struct's fields with values from environment variables.
// It takes a pointer to a struct and optional configuration options.
//
// Each struct field's name is converted to SCREAMING_SNAKE_CASE to determine the
// environment variable name. For example:
//   - Field "ServerPort" looks for "SERVER_PORT"
//   - Field "DatabaseURL" looks for "DATABASE_URL"
//   - Nested field "Database.Password" looks for "DATABASE_PASSWORD"
//
// Supported field types out of the box:
//   - Basic types: string, bool, int*, uint*, float*
//   - time.Duration, time.Time (RFC3339 format)
//   - net.IP, *net.IPNet (CIDR)
//   - *url.URL, *regexp.Regexp
//   - json.RawMessage
//   - []byte (base64 encoded)
//   - Slices of any supported type (comma-separated values)
//
// Features:
//   - Custom type support via WithTypeHandler
//   - Default values via struct tags: `default:"value"`
//   - Custom env names via struct tags: `env:"CUSTOM_NAME"`
//   - Optional prefix for all env vars: WithPrefix("APP")
//   - Nested struct support
//   - Pointer fields are automatically initialized
//
// Example usage:
//
//	type Config struct {
//	    Port        int           `default:"8080"`
//	    Host        string        `env:"SERVICE_HOST"`
//	    Timeout     time.Duration
//	    Database struct {
//	        URL      string
//	        Password string
//	    }
//	}
//
//	var cfg Config
//	err := configloader.Load(&cfg,
//	    WithPrefix("APP"),
//	    WithNameTag("env"),
//	    WithDefaultTag("default"),
//	)
//
// The above will look for these environment variables:
//   - APP_PORT (default: "8080")
//   - SERVICE_HOST (custom name via tag)
//   - APP_TIMEOUT
//   - APP_DATABASE_URL
//   - APP_DATABASE_PASSWORD
//
// Returns an error if:
//   - The value is not a pointer to a struct
//   - Required environment variables are missing
//   - Type conversion fails for any field
//   - Any field has an unsupported type
func Load(value any, opts ...Option) error {
	loader := defaultLoader()
	for _, opt := range opts {
		opt(loader)
	}

	return loader.Load(value)
}

func (l *Loader) Load(val any) (errs error) {
	ptrValue := reflect.ValueOf(val)
	if ptrValue.Kind() != reflect.Pointer {
		return fmt.Errorf("val must be a pointer, got '%s'", reflect.TypeOf(val).String())
	}
	if ptrValue.IsNil() {
		return fmt.Errorf("val cannot be nil")
	}

	// Iterate over each leaf variables of the struct.
	for variable := range traverse(Field{Value: ptrValue}, l.getChildren) {
		value, err := l.lookup(l.keyName(variable), variable)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		err = l.set(value, variable)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
	}
	return errs
}

func (l *Loader) keyName(variable Field) string {
	var names []string
	if l.prefix != "" {
		names = append(names, l.prefix)
	}
	for _, field := range variable.Path {
		if name, found := field.Tag.Lookup(l.nameTag); found {
			// TODO: Does it make sense to replace the whole name?
			names = []string{name}
			break
		}
		names = append(names, l.fieldConversion(field.Name))
	}
	key := strings.Join(names, "_")
	return key
}

func (l *Loader) handler(typ reflect.Type) (func(reflect.Value, string) error, error) {
	_, ok := l.handlers[typ]
	if ok {
		return func(value reflect.Value, val string) error {
			v, err := l.handlers[typ](val)
			if err == nil {
				value.Set(reflect.ValueOf(v))
			}
			return err
		}, nil
	}
	if typ.Kind() == reflect.Slice {
		eh, err := l.handler(typ.Elem())
		if err == nil {
			return func(value reflect.Value, val string) (err error) {
				elements := strings.Split(val, ",")
				slice := reflect.MakeSlice(value.Type(), len(elements), len(elements))
				for i, element := range elements {
					herr := eh(slice.Index(i), element)
					if herr != nil {
						err = errors.Join(err, herr)
					}
				}
				value.Set(slice)
				return err
			}, nil
		}
	}
	return nil, &UnsupportedTypeError{typ}
}

func (l *Loader) canHandle(field reflect.Value) bool {
	if !field.CanSet() || !field.IsValid() {
		return false
	}
	_, err := l.handler(field.Type())
	return err == nil
}

// traverse yields the leaf nodes of a tree.
func traverse[T any](root T, childFn func(T) []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		var traverse func(T) bool
		traverse = func(node T) bool {
			children := childFn(node)
			if len(children) == 0 {
				return yield(node)
			}
			for _, child := range children {
				if !traverse(child) {
					return false
				}
			}
			return true
		}
		traverse(root)
	}
}

func (l *Loader) getChildren(current Field) []Field {
	if l.canHandle(current.Value) {
		return nil
	}

	// Unwrap the current if it's a pointer.
	if current.Value.Kind() == reflect.Ptr {
		if current.Value.IsNil() {
			current.Value.Set(reflect.New(current.Value.Type().Elem()))
		}
		return []Field{{
			Value: current.Value.Elem(),
			Path:  current.Path,
		}}
	}

	if current.Value.Kind() == reflect.Struct {
		var children []Field
		for i := 0; i < current.Value.NumField(); i++ {
			fv := current.Value.Field(i)
			ft := current.Value.Type().Field(i)
			if !fv.CanSet() || !fv.IsValid() {
				continue
			}
			children = append(children, Field{
				Value: fv,
				Path:  append(current.Path, ft),
			})
		}
		return children
	}
	return nil
}

func (l *Loader) lookup(key string, variable Field) (string, error) {
	field := variable.Path[len(variable.Path)-1]
	value, found := l.env(key)
	if !found || value == "" {
		if value, found = field.Tag.Lookup(l.defaultTag); !found {
			return "", MissingEnvErr{Key: key, Field: variable}
		}
	}
	return value, nil
}

func (l *Loader) set(value string, variable Field) error {
	handler, err := l.handler(variable.Value.Type())
	if err != nil {
		return err
	}
	err = handler(variable.Value, value)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	return nil
}
