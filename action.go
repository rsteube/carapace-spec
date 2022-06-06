package spec

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// ActionMacro completes given macro
func ActionMacro(s string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		r := regexp.MustCompile(`^\$(?P<macro>[^(]*)(\((?P<arg>.*)\))?$`)
		if !r.MatchString(s) {
			return carapace.ActionMessage(fmt.Sprintf("malformed macro: '%v'", s))
		}

		matches := findNamedMatches(r, s)
		if m, ok := macros[matches["macro"]]; !ok {
			return carapace.ActionMessage(fmt.Sprintf("unknown macro: '%v'", s))
		} else {
			return m.f(matches["arg"])
		}
	})
}

// ActionSpec completes a spec
func ActionSpec(path string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		abs, err := c.Abs(path)
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}

		content, err := os.ReadFile(abs)
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}

		var cmd Command
		if err := yaml.Unmarshal(content, &cmd); err != nil {
			return carapace.ActionMessage(err.Error())
		}
		return carapace.ActionExecute(cmd.ToCobra())
	})
}

func parseAction(cmd *cobra.Command, arr []string) carapace.Action {
	if !cmd.DisableFlagParsing {
		for _, entry := range arr {
			if strings.HasPrefix(entry, "$") {
				macro := strings.SplitN(strings.TrimPrefix(entry, "$"), "(", 2)[0]
				if m, ok := macros[macro]; ok && m.disableFlagParsing {
					cmd.DisableFlagParsing = true // implicitly disable flag parsing
					break
				}
			}
		}
	}

	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		// TODO yuck - where to set thes best?
		for index, arg := range c.Args {
			c.Setenv(fmt.Sprintf("C_ARG%v", index), arg)
		}
		c.Setenv("C_CALLBACK", c.CallbackValue)

		cmd.Flags().Visit(func(f *pflag.Flag) {
			c.Setenv(fmt.Sprintf("C_FLAG_%v", strings.ToUpper(f.Name)), f.Value.String())
		})

		listDelimiter := ""
		nospace := false
		chdir := ""
		multiparts := ""

		// TODO don't alter the map each time, solve this differently
		addCoreMacro("chdir", MacroI(func(s string) carapace.Action {
			chdir = s
			return carapace.ActionValues()
		}))
		addCoreMacro("list", MacroI(func(s string) carapace.Action {
			listDelimiter = s
			return carapace.ActionValues()
		}))
		addCoreMacro("multiparts", MacroI(func(s string) carapace.Action {
			multiparts = s
			return carapace.ActionValues()
		}))
		addCoreMacro("nospace", MacroI(func(s string) carapace.Action {
			nospace = true
			return carapace.ActionValues()
		}))

		batch := carapace.Batch()
		vals := make([]string, 0)
		for _, elem := range arr {
			if elemSubst, err := c.Envsubst(elem); err != nil {
				batch = append(batch, carapace.ActionMessage(fmt.Sprintf("%v: %v", err.Error(), elem)))
			} else if strings.HasPrefix(elemSubst, "$") { // macro
				batch = append(batch, ActionMacro(elemSubst))
			} else {
				vals = append(vals, parseValue(elemSubst)...)
			}
		}
		batch = append(batch, carapace.ActionStyledValuesDescribed(vals...))

		action := batch.ToA()
		if chdir != "" {
			action = action.Chdir(chdir)
		}
		if multiparts != "" {
			actionCopy := action
			action = carapace.ActionCallback(func(c carapace.Context) carapace.Action {
				return actionCopy.Invoke(c).ToMultiPartsA(multiparts)
			})
		}

		if listDelimiter != "" {
			return carapace.ActionMultiParts(listDelimiter, func(c carapace.Context) carapace.Action {
				for index, arg := range c.Parts {
					c.Setenv(fmt.Sprintf("C_PART%v", index), arg)
				}
				c.Setenv("C_CALLBACK", c.CallbackValue)

				return action.Invoke(c).Filter(c.Parts).ToA()
			})
		} else if nospace {
			return action.NoSpace()
		}
		return action.Invoke(c).ToA()
	})
}

func parseValue(s string) []string {
	if splitted := strings.SplitN(s, "\t", 3); len(splitted) > 2 {
		return splitted
	} else if len(splitted) > 1 {
		return []string{splitted[0], splitted[1], style.Default}
	} else {
		return []string{splitted[0], "", style.Default}
	}
}

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}
