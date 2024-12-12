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

func WithTypeHandler[T any](f func(string) (T, error)) Option {
	return func(l *Loader) {
		var zero T
		l.handlers[reflect.TypeOf(zero)] = func(value string) (any, error) {
			return f(value)
		}
	}
}

// WithNameTag sets the tag used to override environment variable names.
func WithNameTag(tag string) Option {
	return func(l *Loader) {
		l.nameTag = tag
	}
}

// WithDefaultTag sets the tag to use for default values.
func WithDefaultTag(tag string) Option {
	return func(l *Loader) {
		l.defaultTag = tag
	}
}

// WithPrefix sets the prefix for environment variable names.
func WithPrefix(prefix string) Option {
	return func(l *Loader) {
		l.prefix = prefix
	}
}

// WithEnv sets a custom environment variable lookup function.
func WithEnv(env func(string) (string, bool)) Option {
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
