package annotation

import (
	"bytes"
	"fmt"
	"github.com/expgo/ag/api"
	"github.com/expgo/generic/stream"
	"io"
	"strings"
)

type Generator struct {
	singletons []*Singleton
}

func (g *Generator) GetImports() []string {
	return []string{"github.com/expgo/factory"}
}

func (g *Generator) WriteConst(wr io.Writer) error {
	return nil
}

func (g *Generator) WriteInitFunc(wr io.Writer) error {
	buf := bytes.NewBuffer([]byte{})

	buf.WriteString("func init() {")

	for _, s := range g.singletons {
		if s.NamedOnly {
			if len(s.Name) == 0 {
				return fmt.Errorf("%s's Singleton annotation must with name param", s.typeName)
			}
			buf.WriteString(fmt.Sprintf(`factory.NamedSingleton[%s]("%s")`, s.typeName, s.Name))
		} else {
			buf.WriteString(fmt.Sprintf(`factory.Singleton[%s]()`, s.typeName))
			if len(s.Name) > 0 {
				buf.WriteString(fmt.Sprintf(`.Name("%s")`, s.Name))
			}
		}

		if s.UseConstructor {
			buf.WriteString(".UseConstructor(true)")
		}

		if len(s.InitMethod) > 0 {
			buf.WriteString(fmt.Sprintf(`.InitMethodName("%s")`, s.InitMethod))
		}

		if len(s.Init) > 0 {
			var quoted []string

			for _, v := range s.Init {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, v))
			}

			buf.WriteString(fmt.Sprintf(`.InitParams(%s)`, strings.Join(quoted, ",")))
		}

		buf.WriteString("\n")
	}

	buf.WriteString("}")

	_, err := io.Copy(wr, buf)
	return err
}

func (g *Generator) WriteBody(wr io.Writer) error {
	return nil
}

func newGenerator(singletons []*Singleton) (api.Generator, error) {
	sorted := stream.Must(stream.Of(singletons).Sort(func(x, y *Singleton) int { return strings.Compare(x.typeName, y.typeName) }).ToSlice())
	return &Generator{sorted}, nil
}
