package annotation

import (
	"bytes"
	"fmt"
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"io"
	"sort"
	"strings"
)

type PluginGenerator struct {
	singletons []*Singleton
	factories  []*Factory
}

func (g *PluginGenerator) GetImports() []string {
	return []string{"github.com/expgo/factory"}
}

func (g *PluginGenerator) WriteConst(wr io.Writer) error {
	anyMatch := false
	for _, s := range g.singletons {
		if s.LocalVar || s.LocalGetter {
			anyMatch = true
			break
		}
	}

	if anyMatch {
		buf := bytes.NewBuffer([]byte{})

		buf.WriteString("var(\n")

		for _, s := range g.singletons {
			if s.LocalVar || s.LocalGetter {
				if err := s.WriteString(buf); err != nil {
					return err
				}
			}
		}

		buf.WriteString(")\n")

		_, err := io.Copy(wr, buf)
		return err
	}

	return nil
}

func (g *PluginGenerator) WriteInitFunc(wr io.Writer) error {
	anyMatch := false
	for _, s := range g.singletons {
		if !(s.LocalVar || s.LocalGetter) {
			anyMatch = true
			break
		}
	}

	if anyMatch || len(g.factories) > 0 {
		buf := bytes.NewBuffer([]byte{})

		buf.WriteString("func init() {\n")

		for _, f := range g.factories {
			if f.isFunc {
				buf.WriteString(fmt.Sprintf(`factory.Factory[%s](%s)`, f.funcReturn, f.funcName))
			} else {
				buf.WriteString(fmt.Sprintf(`factory.Factory[%s](factory.New[%s]())`, f.funcReturn, f.structName))
				if f.funcName != factory.NewMethodName {
					buf.WriteString(fmt.Sprintf(`.MethodName("%s")`, f.funcName))
				}
			}

			if len(f.Params) > 0 {
				var quoted []string

				for _, p := range f.Params {
					quoted = append(quoted, fmt.Sprintf(`"%s"`, p))
				}

				buf.WriteString(fmt.Sprintf(`.Params(%s)`, strings.Join(quoted, ",")))
			}

			buf.WriteString(".CheckValid()\n")
		}

		for _, s := range g.singletons {
			if !(s.LocalVar || s.LocalGetter) {
				if err := s.WriteString(buf); err != nil {
					return err
				}
			}
		}

		buf.WriteString("}\n")

		_, err := io.Copy(wr, buf)
		return err
	}

	return nil
}

func (g *PluginGenerator) WriteBody(wr io.Writer) error {
	return nil
}

func newGenerator(singletons []*Singleton, factories []*Factory) (api.Generator, error) {
	sortedSingletons := append([]*Singleton(nil), singletons...)
	sort.Slice(sortedSingletons, func(i, j int) bool {
		return strings.Compare(sortedSingletons[i].typeName, sortedSingletons[j].typeName) < 0
	})

	sortedFactories := append([]*Factory(nil), factories...)
	sort.Slice(sortedFactories, func(i, j int) bool {
		return strings.Compare(sortedFactories[i].funcName, sortedFactories[j].funcName) < 0
	})

	return &PluginGenerator{sortedSingletons, sortedFactories}, nil
}
