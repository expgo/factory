package annotation

import (
	"bytes"
	"github.com/expgo/ag/api"
	"github.com/expgo/generic/stream"
	"io"
	"strings"
)

type PluginGenerator struct {
	singletons []*Singleton
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
	if stream.Must(stream.Of(g.singletons).AnyMatch(func(s *Singleton) (bool, error) { return !s.LocalVar, nil })) {
		buf := bytes.NewBuffer([]byte{})

		buf.WriteString("func init() {\n")

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

func newGenerator(singletons []*Singleton) (api.Generator, error) {
	sorted := stream.Must(stream.Of(singletons).Sort(func(x, y *Singleton) int { return strings.Compare(x.typeName, y.typeName) }).ToSlice())
	return &PluginGenerator{sorted}, nil
}
