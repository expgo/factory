package annotation

import (
	"bytes"
	"fmt"
	"github.com/expgo/ag/api"
	"github.com/expgo/generic/stream"
	"io"
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
	if stream.Must(stream.Of(g.singletons).AnyMatch(func(s *Singleton) (bool, error) { return s.LocalVar, nil })) {

		buf := bytes.NewBuffer([]byte{})

		buf.WriteString("var(\n")

		for _, s := range g.singletons {
			if s.LocalVar {
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
	if stream.Must(stream.Of(g.singletons).AnyMatch(func(s *Singleton) (bool, error) { return !s.LocalVar, nil })) ||
		len(g.factories) > 0 {
		buf := bytes.NewBuffer([]byte{})

		buf.WriteString("func init() {\n")

		for _, f := range g.factories {
			buf.WriteString(fmt.Sprintf(`factory.Factory[%s](%s)`, f.funcReturn, f.funcName) + "\n")
		}

		for _, s := range g.singletons {
			if !s.LocalVar {
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
	sortedSingletons := stream.Must(stream.Of(singletons).Sort(func(x, y *Singleton) int { return strings.Compare(x.typeName, y.typeName) }).ToSlice())
	sortedFactories := stream.Must(stream.Of(factories).Sort(func(x, y *Factory) int { return strings.Compare(x.funcName, y.funcName) }).ToSlice())
	return &PluginGenerator{sortedSingletons, sortedFactories}, nil
}
