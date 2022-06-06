package spec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rsteube/carapace"
	"gopkg.in/yaml.v3"
)

type Macro struct {
	f                  func(string) carapace.Action
	s                  func() string
	disableFlagParsing bool
}

func (m Macro) Signature() string { return m.s() }
func (m Macro) NoFlag() Macro     { m.disableFlagParsing = true; return m }

var macros = make(map[string]Macro)

func addCoreMacro(s string, m Macro) {
	macros[s] = m
}

// AddMacro adds a custom macro
func AddMacro(s string, m Macro) {
	macros["_"+s] = m
}

// MacroN careates a macro without an argument
func MacroN(f func() carapace.Action) Macro {
	return Macro{
		f: func(s string) carapace.Action {
			return f()
		},
		s: func() string { return "" },
	}
}

// MacroI careates a macro with an argument
func MacroI[T any](f func(t T) carapace.Action) Macro {
	return Macro{
		f: func(s string) carapace.Action {
			var t T
			if reflect.TypeOf(t).Kind() == reflect.String {
				reflect.ValueOf(&t).Elem().SetString(s)
			} else {
				if err := yaml.Unmarshal([]byte(s), &t); err != nil {
					return carapace.ActionMessage(err.Error())
				}

				if s == "" {
					if m := reflect.ValueOf(&t).MethodByName("Default"); m.IsValid() && m.Type().NumIn() == 0 {
						values := m.Call([]reflect.Value{}) // TODO check if needs args
						if len(values) > 0 && values[0].Type().AssignableTo(reflect.TypeOf(t)) {
							reflect.ValueOf(&t).Elem().Set(values[0])
						}
					}

				}
			}
			return f(t)
		},
		s: func() string { return signature(new(T)) },
	}
}

// MacroV careates a macro with a variable argument
func MacroV[T any](f func(s ...T) carapace.Action) Macro {
	return Macro{
		f: func(s string) carapace.Action {
			if s == "" {
				return f()
			}

			var t []T
			if err := yaml.Unmarshal([]byte(s), &t); err != nil {
				return carapace.ActionMessage(fmt.Sprintf("malformed macro arg: '%v', expected '%v'", s, reflect.TypeOf(t)))
			}
			return f(t...)
		},
		s: func() string { return fmt.Sprintf("[%v]", signature(new(T))) },
	}
}

func signature(i interface{}) string {
	elem := reflect.ValueOf(i).Elem()
	switch elem.Kind() {
	case reflect.Struct:
		out, err := yaml.Marshal(i)
		if err != nil {
			return err.Error()
		}
		lines := strings.Split(string(out), "\n")
		return fmt.Sprintf("{%v}", strings.Join(lines[:len(lines)-1], ", "))

	case reflect.Slice:
		ptr := reflect.New(elem.Type().Elem()).Interface()
		return fmt.Sprintf("[%v]", signature(ptr))

	case reflect.String:
		return `""`

	default:
		return fmt.Sprintf("%v", reflect.ValueOf(i).Elem())
	}
}
