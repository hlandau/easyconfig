// Package adaptflag adapts registered configurables to common flag parsing
// packages, thereby making configurables configurable from the command line.
package adaptflag

import "fmt"
import "flag"
import "github.com/ogier/pflag"
import "gopkg.in/alecthomas/kingpin/v2"
import "gopkg.in/hlandau/configurable.v1"
import "strings"

var shortFlags = map[string]rune{}
var shortFlagsReverse = map[rune]string{}

func MapShort(name string, s rune) {
	shortFlags[name] = s
	if _, ok := shortFlagsReverse[s]; ok {
		panic(fmt.Sprintf("short flag already mapped: %#v=%#v", name, s))
	}
	shortFlagsReverse[s] = name
}

func name(c configurable.Configurable) (name string, ok bool) {
	v, ok := c.(interface {
		CfName() string
	})
	if !ok {
		return
	}

	return v.CfName(), true
}

func usageSummaryLine(c configurable.Configurable) (s string, ok bool) {
	v, ok := c.(interface {
		CfUsageSummaryLine() string
	})
	if !ok {
		return
	}

	return v.CfUsageSummaryLine(), true
}

func defaultValue(c configurable.Configurable) (dflt interface{}, ok bool) {
	v, ok := c.(interface {
		CfDefaultValue() interface{}
	})
	if !ok {
		return nil, false
	}

	return v.CfDefaultValue(), true
}

var errNotSupported = fmt.Errorf("not supported")

type value struct {
	c configurable.Configurable
}

// The flag package uses this to get the default value.
func (v *value) String() string {
	dflt, ok := defaultValue(v.c)
	if !ok {
		return "[configurable]"
	}

	return fmt.Sprintf("%#v", dflt)
}

func (v *value) Set(s string) error {
	cs, ok := v.c.(interface {
		CfSetValue(v interface{}) error
	})
	if !ok {
		return errNotSupported
	}

	cp, ok := v.c.(interface {
		CfGetPriority() configurable.Priority
		CfSetPriority(priority configurable.Priority)
	})
	if !ok {
		return cs.CfSetValue(s)
	}

	if cp.CfGetPriority() <= configurable.FlagPriority {
		err := cs.CfSetValue(s)
		if err != nil {
			return err
		}

		cp.CfSetPriority(configurable.FlagPriority)
	}

	return nil
}

func (v *value) Get() interface{} {
	cg, ok := v.c.(interface {
		CfValue() interface{}
	})
	if !ok {
		return nil // ...
	}

	return cg.CfValue()
}

func (v *value) IsBoolFlag() bool {
	_, ok := v.Get().(bool)
	return ok
}

var adapted = map[interface{}]struct{}{}

func adapt(path []string, c configurable.Configurable, f AdaptFunc) error {
	_, ok := adapted[c]
	if ok {
		return nil
	}

	name, ok := name(c)
	if !ok {
		return errNotSupported
	}

	_, ok = c.(interface {
		CfSetValue(v interface{}) error
	})
	if !ok {
		return errNotSupported
	}

	v := &value{c: c}
	usage, _ := usageSummaryLine(c)

	dfltv, ok := defaultValue(c)
	dfltstr := ""
	if ok {
		dfltstr = fmt.Sprintf("%v", dfltv)
	}

	f(Info{
		Name:               name,
		Usage:              usage,
		Value:              v,
		Path:               path,
		DefaultValueString: dfltstr,
	})

	adapted[c] = struct{}{}
	return nil
}

// Gathered information about a configurable. This information makes it easy to
// call flag.Var-like functions.
type Info struct {
	Name               string
	Usage              string
	Path               []string
	Value              Value
	DefaultValueString string
}

// Called repeatedly by AdoptWithFunc. Your implementation of this function
// should register the Value with the details provided. It is especially
// suitable for use with functions like flag.Var or packages which provide
// similar interfaces.
type AdaptFunc func(info Info)

func recursiveAdapt(path []string, c configurable.Configurable, f AdaptFunc) error {
	adapt(path, c, f)
	cc, ok := c.(interface {
		CfChildren() []configurable.Configurable
	})
	if ok {
		n, ok := name(c)
		if ok {
			p := make([]string, 0, len(path)+1)
			path = append(p, path...)
			path = append(p, n)
		}
		for _, ch := range cc.CfChildren() {
			recursiveAdapt(path, ch, f)
		}
	}
	return nil
}

// The interface which this package exposes to the flag packages it adapts to.
type Value interface {
	String() string
	Set(x string) error
}

// Similar to Adapt, but allows you to register to the flag package of your
// choice, so long as it implements an interface similar to the flag.Var
// function.
func AdaptWithFunc(f AdaptFunc) {
	configurable.Visit(func(c configurable.Configurable) error {
		return recursiveAdapt(nil, c, f)
	})
}

// Adapt registers all registered configurables as flags with the flag and
// ogier/pflag packages. Note that Adapt will not do anything with
// Configurables which it has already adapted once, thus it is safe to call
// this function multiple times.
func Adapt() {
	AdaptWithFunc(func(info Info) {
		dpn := DottedPath(info.Path)
		if len(dpn) > 0 {
			dpn += "."
		}
		dpn += info.Name
		flag.Var(info.Value, dpn, info.Usage)
		pflag.Var(info.Value, dpn, info.Usage)
		fl := kingpin.Flag(dpn, info.Usage)
		if info.DefaultValueString != "" {
			fl = fl.PlaceHolder(info.DefaultValueString)
		} else {
			fl = fl.PlaceHolder("\"\"")
		}
		if r, ok := shortFlags[dpn]; ok {
			fl = fl.Short(r)
		}
		fl.SetValue(info.Value)
	})
}

func DottedPath(path []string) string {
	return strings.Join(path, ".")
}
