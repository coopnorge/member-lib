package configloader

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type Loader struct {
	prefix, tag string
	fieldNamer  func(string) string
	env         func(string) string
}

func defaultLoader() *Loader {
	return &Loader{
		fieldNamer: strcase.ToScreamingSnake,
		env:        os.Getenv,
	}
}

type Option func(*Loader)

// WithPrefix sets the prefix for environment variable names.
func WithTag(tag string) Option {
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

// WithFieldNamer sets a custom field naming function.
func WithFieldNamer(namer func(string) string) Option {
	return func(l *Loader) {
		l.fieldNamer = namer
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

	// reflect.Type -> Pass it to load
	return loader.Load(value)
}

// Loader loads a configuration struct from environment variables.
// It supports nested structs and handles type conversion for basic types.
// TODO: Remove params and replace with the With Pattern.
func (l *Loader) Load(val any) error {
	ptrValue := reflect.ValueOf(val)
	if ptrValue.Kind() != reflect.Pointer {
		return fmt.Errorf("need a pointer to load values into. Got %T", val)
	}
	v := ptrValue.Elem()
	println(v.String())
	t := ptrValue.Type().Elem() // Get the type from the pointer type
	println(t.String())

	if err := l.loadFields(v, t, l.prefix); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	return nil
}

// loadFields recursively processes struct fields and loads values from environment variables.
func (l *Loader) loadFields(v reflect.Value, t reflect.Type, prefix string) error {
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct {
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "_"
			}
			newPrefix += l.fieldNamer(fieldType.Name)

			if err := l.loadFields(field, fieldType.Type, newPrefix); err != nil {
				return fmt.Errorf("error loading nested struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		// Get environment variable name from tag or field name
		var found bool
		envName := fieldType.Tag.Get(l.tag)
		if envName == "" {
			found = false
			envName = l.fieldNamer(fieldType.Name)
		} else {
			found = true
		}

		// Add prefix to environment variable name
		// Only add prefix if the user didn't set env var directly.
		if prefix != "" && !found {
			envName = prefix + "_" + envName
		}

		// Get environment variable value
		slog.Info("attempting to load field", slog.String("field_name", envName))
		envValue := l.env(envName)
		if envValue == "" {
			// Check if the field has a default value set
			if field.IsZero() {
				return fmt.Errorf("environment variable not set: %s", envName)
			}
			continue // Keep default value if env var is not set
		}

		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue converts string value from environment variable to appropriate field type.
func setFieldValue(field reflect.Value, value string) error {
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
			return err
		}
		field.SetBool(val)
	case reflect.Slice:
		// Handle slice types (split by comma)
		elements := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(elements), len(elements))
		for i, element := range elements {
			if err := setFieldValue(slice.Index(i), strings.TrimSpace(element)); err != nil {
				return err
			}
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
