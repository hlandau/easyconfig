// Package cstruct allows for the automatic generation of configurables from an
// annotated structure.
//
// To use cstruct, you call New or MustNew, passing a pointer to an instance of
// an annotated structure type.
//
// The supported field types are string, int and bool. A field is only used if
// it is public and has the `default` or `usage` tags specified on it, or both.
// The name of the field will be used as the configurable name.
//
// The following tags can be placed on fields:
//
//   default: The default value as a string.
//   usage: A one-line usage summary.
//
// Once you have created a cstruct Configurable group, you must register it
// appropriately as you see fit, for example by calling configurable.Register.
package cstruct

import "time"
import "fmt"
import "reflect"
import "strings"
import "regexp"
import "strconv"
import "gopkg.in/hlandau/configurable.v1"

type group struct {
	configurables []configurable.Configurable
	name          string
}

func (g *group) CfChildren() []configurable.Configurable {
	return g.configurables
}

func (g *group) CfName() string {
	return g.name
}

type value struct {
	name, usageSummaryLine, envVarName string
	v                                  reflect.Value
	defaultValue                       interface{}
	priority                           configurable.Priority
}

func (v *value) CfName() string {
	return v.name
}

func (v *value) String() string {
	return fmt.Sprintf("cstruct-value(%s)", v.CfName())
}

func (v *value) CfGetValue() interface{} {
	return v.v.Interface()
}

func (v *value) CfDefaultValue() interface{} {
	return v.defaultValue
}

func (v *value) CfUsageSummaryLine() string {
	return v.usageSummaryLine
}

func (v *value) CfEnvVarName() string {
	return v.envVarName
}

func (v *value) CfGetPriority() configurable.Priority {
	return v.priority
}

func (v *value) CfSetPriority(priority configurable.Priority) {
	v.priority = priority
}

// Like New, but panics on failure.
func MustNew(target interface{}, name string) (c configurable.Configurable) {
	c, err := New(target, name)
	if err != nil {
		panic(err)
	}

	return c
}

// Creates a new group Configurable, with children representing the fields.
//
// The Configurables set the values of the fields of the instance.
func New(target interface{}, name string) (c configurable.Configurable, err error) {
	t := reflect.TypeOf(target)
	v := reflect.ValueOf(target)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}

	if t.Kind() != reflect.Struct {
		err = fmt.Errorf("target interface is not a struct: %v", t)
		return
	}

	g := &group{
		name: name,
	}
	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		usage := field.Tag.Get("usage")
		dflt := field.Tag.Get("default")
		envVarName := field.Tag.Get("env")

		if usage == "" && dflt == "" {
			continue
		}

		vf := v.FieldByIndex(field.Index)

		if !vf.CanSet() {
			err = fmt.Errorf("field not assignable")
			return
		}

		vv := &value{
			v:                vf,
			name:             name,
			envVarName:       envVarName,
			usageSummaryLine: usage,
		}

		if dflt != "" {
			var dfltv reflect.Value
			dfltv, err = parseString(dflt, vf.Type())
			if err != nil {
				err = fmt.Errorf("invalid default value: %#v: %v", dflt, err)
				return
			}

			vv.defaultValue = dfltv.Interface()
			err = vv.CfSetValue(dfltv.Interface())
			if err != nil {
				panic(fmt.Sprintf("cannot set default value on field: %v", err))
			}
		}

		g.configurables = append(g.configurables, vv)

		// Do the type check now
		/*switch field.Type.Kind() {
		case reflect.Int:
		case reflect.String:
		case reflect.Bool:
		default:
			err = fmt.Errorf("unsupported field type: %v", field.Type)
			return
		}*/
	}

	return g, nil
}

func (v *value) CfSetValue(nw interface{}) error {
	return coercingSet(v.v, reflect.ValueOf(nw))
}

// Sets a field value to a new value, coercing the new value if necessary.
// Returns an error if coersion is impossible.
//
// Setting a slice field to a non-slice value results in the new value getting
// appended to the existing values. (Although this isn't useful if you're
// dealing with slices of slices.)
func coercingSet(field reflect.Value, newValue reflect.Value) error {
	coerced, err := coerceValue(newValue, &field, field.Type())
	if err != nil {
		return err
	}

	field.Set(coerced)
	return nil
}

// Ensures that value is assignable to targetType, converting it if necessary.
//
// If value is already assignable to targetType, returns value.
// If value is not already assignable to targetType, constructs a new
// reflect.Value that is and returns it, or, failing that, returns an error.
func coerceValue(value reflect.Value, oldValue *reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	if value.Type().AssignableTo(targetType) {
		return value, nil
	}

	// Ensure that []interface{} (from e.g. TOML) can be converted to []T for some T.
	if value.Type().Kind() == reflect.Slice && targetType.Kind() == reflect.Slice {
		slice := reflect.MakeSlice(targetType, 0, 0)

		for i := 0; i < value.Len(); i++ {
			cv, err := coerceValue(value.Index(i), nil, targetType.Elem())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("element of slice %#v cannot be coerced to type %v", value.Index(i).Interface(), targetType.Elem())
			}

			slice = reflect.Append(slice, cv)
		}

		return slice, nil
	}

	// The target is a slice but the source isn't, and we have a previous value
	// we can accumulate from.
	if targetType.Kind() == reflect.Slice && oldValue != nil {
		cv, err := coerceValue(value, nil, targetType.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("value %#v cannot be coerced to type %v", value, targetType.Elem())
		}

		return reflect.Append(*oldValue, cv), nil
	}

	// Parse string.
	if value.Type().Kind() == reflect.String {
		return parseString(value.String(), targetType)
	}

	// Don't know how to coerce.
	return reflect.Value{}, fmt.Errorf("don't know how to coerce %v (%v) to type %v", value.String(), value.Type(), targetType)
}

var re_no = regexp.MustCompile(`(?i)(00*|no?|f(alse)?)`)

// Tries to coerce a string to the specified type.
func parseString(s string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Int:
		n, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(int(n)), nil

	case reflect.Bool:
		on := (s != "" && !re_no.MatchString(s))

		return reflect.ValueOf(on), nil

	case reflect.String:
		return reflect.ValueOf(s), nil

	case reflect.Int64:
		if t.String() == "time.Duration" {
			d, err := time.ParseDuration(s)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(d), nil
		}
		break

	default:
		break
	}

	return reflect.Value{}, fmt.Errorf("cannot coerce string %#v to type %v (%v)", s, t, t.Kind())
}
