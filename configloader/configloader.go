package configloader

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Loader struct {
	prefix, tag string

	// Type of conversion the struct field will take. Default is SnakeCase.
	fieldConversion func(string) string

	// Custom GetEnv function.
	env func(string) string

	// Map of type handlers where key is reflect.Type and value is the handler function
	handlers map[reflect.Type]func(string) (any, error)
}

func defaultLoader() *Loader {
	return &Loader{
		handlers:        make(map[reflect.Type]func(string) (any, error), 0),
		fieldConversion: strcase.ToScreamingSnake,
		env:             os.Getenv,
	}
}

type Option func(*Loader)

func WithTypeHandler[T any](f func(string) (T, error)) Option {
	return func(l *Loader) {
		var zero T
		l.handlers[reflect.TypeOf(zero)] = func(value string) (any, error) {
			return f(value)
		}
	}
}

// WithNameTag sets the prefix for environment variable names.
func WithNameTag(tag string) Option {
	return func(l *Loader) {
		l.tag = tag
	}
}

// WithPrefix sets the prefix for environment variable names.
func WithPrefix(prefix string) Option {
	return func(l *Loader) {
		l.prefix = prefix
	}
}

// WithEnv sets a custom environment variable lookup function.
func WithEnv(env func(string) string) Option {
	return func(l *Loader) {
		l.env = env
	}
}

func Load(value any, opts ...Option) error {
	loader := defaultLoader()
	for _, opt := range opts {
		opt(loader)
	}

	return loader.Load(value)
}

// Load loads a configuration struct from environment variables.
// It supports nested structs and handles type conversion for basic types.
func (l *Loader) Load(val any) error {
	ptrValue := reflect.ValueOf(val)
	if ptrValue.Kind() != reflect.Pointer || ptrValue.IsNil() {
		return fmt.Errorf("need a pointer to load values into. Got %s", reflect.TypeOf(val).String())
	}
	v := ptrValue.Elem()
	t := ptrValue.Type().Elem()

	if err := l.loadFields(v, t, l.prefix); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	return nil
}

// loadFields recursively processes struct fields and loads values from environment variables.
// nolint:gocognit // Necessary complexity. High IQ function.
func (l *Loader) loadFields(v reflect.Value, t reflect.Type, prefix string) error {
	if !v.IsValid() {
		return fmt.Errorf("struct should be an initialized pointer")
	}

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() || !field.IsValid() {
			continue
		}
		newPrefix := prefix
		if newPrefix != "" {
			newPrefix += "_"
		}
		convertedFName := l.fieldConversion(fieldType.Name)

		// Handle pointers to structs
		if field.Kind() == reflect.Ptr {
			// Initialize if nil
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			// If it's a pointer to struct, process it
			if field.Elem().Kind() == reflect.Struct {
				newPrefix += convertedFName
				if err := l.loadFields(field.Elem(), fieldType.Type.Elem(), newPrefix); err != nil {
					return fmt.Errorf("error loading nested pointer struct %s: %w", fieldType.Name, err)
				}
				continue
			}
		}

		if field.Kind() == reflect.Struct {
			newPrefix += convertedFName
			if err := l.loadFields(field, fieldType.Type, newPrefix); err != nil {
				return fmt.Errorf("error loading nested struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		var found bool
		envName := fieldType.Tag.Get(l.tag)
		if envName == "" {
			found = false
			envName = convertedFName
		} else {
			envName = strings.Split(envName, ",")[0]
			found = true
		}

		if prefix != "" && !found {
			envName = prefix + "_" + envName
		}

		if err := l.setFieldValue(field, l.env(envName)); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
		}
	}
	return nil
}

// setFieldValue converts string value from environment variable to appropriate field type.
// nolint:funlen // Switch statement makes this a lengthy func.
func (l *Loader) setFieldValue(field reflect.Value, value string) error {
	// skip if value is nil or empty
	if value == "" {
		return nil
	}

	// Override type if a specific handler was given.
	// NOTE: Overrides of Complex Types masks could be picked up instead of the masked type.
	f, ok := l.handlers[field.Type()]
	if ok {
		v, err := f(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(v))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			// Let user know that Parsing is wrong
			return err
		}
		field.SetBool(val)
	case reflect.Slice:
		// Handle slice types (split by comma)
		elements := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(elements), len(elements))
		for i, element := range elements {
			if err := l.setFieldValue(slice.Index(i), strings.TrimSpace(element)); err != nil {
				return err
			}
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
