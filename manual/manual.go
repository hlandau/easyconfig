// Package manual allows you to easily set configurables programmatically.
package manual

import "gopkg.in/hlandau/configurable.v1"
import "strings"
import "fmt"

func ByName(name string) configurable.Configurable {
	parts := strings.Split(name, ".")
	top := topByName(parts[0])
	if top == nil {
		return nil
	}

	return byName(top, parts[1:])
}

func byName(c configurable.Configurable, p []string) configurable.Configurable {
	if len(p) == 0 {
		return c
	}

	cc, ok := c.(interface {
		CfChildren() []configurable.Configurable
	})
	if ok {
		for _, ch := range cc.CfChildren() {
			n, ok := getName(ch)
			if ok && n == p[0] {
				return byName(ch, p[1:])
			}
		}
	}

	return nil
}

var errStop = fmt.Errorf("stop")
var errNotSupported = fmt.Errorf("not supported")

func topByName(name string) configurable.Configurable {
	var cc configurable.Configurable
	configurable.Visit(func(c configurable.Configurable) error {
		n, ok := getName(c)
		if ok && n == name {
			cc = c
			return errStop
		}

		return nil
	})

	return cc
}

func Set(name string, value interface{}) error {
	c := ByName(name)
	if c == nil {
		return fmt.Errorf("configurable not found: %s", name)
	}

	return set(c, value)
}

func set(c configurable.Configurable, s interface{}) error {
	cs, ok := c.(interface {
		CfSetValue(v interface{}) error
	})
	if !ok {
		return errNotSupported
	}

	cp, ok := c.(interface {
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

func getName(c configurable.Configurable) (name string, ok bool) {
	v, ok := c.(interface {
		CfName() string
	})
	if !ok {
		return
	}

	return v.CfName(), true
}
