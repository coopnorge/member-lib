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
	prefix, nameTag, envTag, defaultTag string

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

func (f *Field) Last() reflect.StructField {
	return f.Path[len(f.Path)-1]
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

func defaultLoader() *Loader {
	l := &Loader{
		handlers:        make(map[reflect.Type]func(string) (any, error), 0),
		fieldConversion: strcase.ToScreamingSnake,
		env: func(val string) (string, bool) {
			return os.LookupEnv(strings.ToUpper(val))
		},
		defaultTag: "default",
	}
	for _, opt := range defaultTypeHandlers {
		opt(l)
	}
	return l
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

func (l *Loader) Load(val any) error {
	ptrValue := reflect.ValueOf(val)
	if ptrValue.Kind() != reflect.Pointer {
		return fmt.Errorf("val must be a pointer, got '%s'", reflect.TypeOf(val).String())
	}
	if ptrValue.IsNil() {
		return fmt.Errorf("val cannot be nil")
	}

	errs := ConfigLoadError{
		Value: ptrValue.Elem().Type(),
	}

	// Iterate over each leaf variables of the struct.
	for variable := range traverse(Field{Value: ptrValue}, l.getChildren) {
		err := l.parse(variable)
		if err != nil {
			errs.Add(FieldError{
				Field: variable,
				Err:   err,
			})
		}
	}
	if len(errs.Errors) > 0 {
		return &errs
	}
	return nil
}

func (l *Loader) parse(variable Field) error {
	name, err := l.keyName(variable)
	if err != nil {
		return err
	}
	value, err := l.lookup(name, variable)
	if err != nil {
		return err
	}
	err = l.set(value, variable)
	if err != nil {
		return err
	}
	return nil
}

func (l *Loader) keyName(variable Field) (string, error) {
	var names []string
	if l.prefix != "" {
		names = append(names, l.prefix)
	}

	// Env not last.
	var fieldsWithEnv []reflect.StructField
	for _, field := range variable.Path[:len(variable.Path)-1] {
		if _, found := field.Tag.Lookup(l.envTag); found {
			fieldsWithEnv = append(fieldsWithEnv, field)
		}
	}
	if len(fieldsWithEnv) > 0 {
		return "", fmt.Errorf("´%s´ tag need to be at the end: %v", l.envTag, fieldsWithEnv)
	}

	if name, found := variable.Last().Tag.Lookup(l.envTag); found {
		return name, nil
	}

	for _, field := range variable.Path {
		if value, found := field.Tag.Lookup(l.nameTag); found {
			names = append(names, value)
		} else {
			names = append(names, l.fieldConversion(field.Name))
		}
	}
	key := strings.Join(names, "_")
	return key, nil
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
			if ft.Anonymous {
				children = append(children, Field{
					Value: fv,
					Path:  current.Path,
				})
			} else {
				children = append(children, Field{
					Value: fv,
					Path:  append(current.Path, ft),
				})
			}
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
			return "", MissingEnvError{
				Key: key,
			}
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
		return err
	}
	return nil
}
