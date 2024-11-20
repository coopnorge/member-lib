package configloader

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const defaultTag = "env"

type PreProcessingGetEnvFunc func(string) string

// Loader loads a configuration struct from environment variables.
// It supports nested structs and handles type conversion for basic types.
// TODO: Remove params and replace with the With Pattern
func Loader[T any](customTag string, prefix string, f PreProcessingGetEnvFunc) (*T, error) {
	var config T
	v := reflect.ValueOf(&config).Elem()
	t := v.Type()

	if err := loadFields(v, t, prefix, customTag, f); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &config, nil
}

// loadFields recursively processes struct fields and loads values from environment variables
// TODO: Remove params and replace with the With Pattern
func loadFields(v reflect.Value, t reflect.Type, prefix string, customTag string, f PreProcessingGetEnvFunc) error {
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
			newPrefix += convertNameIntoEnvNotation(fieldType.Name)

			if err := loadFields(field, fieldType.Type, newPrefix, customTag, f); err != nil {
				return fmt.Errorf("error loading nested struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		// Check if customTag is empty
		var tag string
		if customTag != "" {
			tag = customTag
		} else {
			tag = defaultTag
		}

		// Get environment variable name from tag or field name
		var found bool
		envName := fieldType.Tag.Get(f(tag))
		if envName == "" {
			found = false
			envName = convertNameIntoEnvNotation(fieldType.Name)
		} else {
			found = true
		}

		// Add prefix to environment variable name
		// Only add prefix if the user didn't set env var directly.
		if prefix != "" && !found {
			envName = prefix + "_" + envName
		}

		// Get environment variable value
		// TODO: Omitempty? Standard across go
		slog.Info("attempting to load field", slog.String("field_name", envName))
		envValue := os.Getenv(envName)
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

// setFieldValue converts string value from environment variable to appropriate field type
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

// convertNameIntoEnvNotation converts a struct field name into an environment variable notation
// e.g., DatabaseURL -> DATABASE_URL, OAuth2Token -> OAUTH2_TOKEN
func convertNameIntoEnvNotation(name string) string {
	if name == "" {
		return ""
	}

	var result strings.Builder
	runes := []rune(name)

	// Handle first character
	result.WriteRune(unicode.ToUpper(runes[0]))

	for i := 1; i < len(runes); i++ {
		current := runes[i]

		// Look ahead and behind
		var next rune
		if i+1 < len(runes) {
			next = runes[i+1]
		}
		prev := runes[i-1]

		// Conditions for adding underscore:
		// 1. Current is uppercase and previous is lowercase (camelCase -> CAMEL_CASE)
		// 2. Current is uppercase and next exists and is lowercase (URLEnd -> URL_END)
		// 3. Current is number and previous is letter (OAuth2 -> OAUTH2)
		// 4. Previous is number and current is uppercase (API2Token -> API2_TOKEN)
		needsUnderscore := false

		if unicode.IsUpper(current) {
			// Handle transition from acronym to new word
			// e.g., APIConfig -> API_CONFIG
			if unicode.IsUpper(prev) && next != 0 && unicode.IsLower(next) {
				needsUnderscore = true
			}

			// Handle transition from lowercase to uppercase
			// e.g., someAPI -> SOME_API
			if unicode.IsLower(prev) {
				needsUnderscore = true
			}
		}

		// Handle numbers
		if unicode.IsNumber(current) && unicode.IsLetter(prev) {
			if !unicode.IsNumber(prev) {
				needsUnderscore = true
			}
		}
		if unicode.IsLetter(current) && unicode.IsNumber(prev) {
			if unicode.IsUpper(current) {
				needsUnderscore = true
			}
		}

		if needsUnderscore {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToUpper(current))
	}

	return result.String()
}

// NoPreProcessing is a noop operation
func NoPreProcessing(s string) string {
	return s
}
