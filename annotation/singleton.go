package annotation

import (
	"fmt"
	"io"
	"strings"
)

type Singleton struct {
	Name           string
	NamedOnly      bool   `value:"false"`
	UseConstructor bool   `value:"false"`
	LazyInit       bool   `value:"true"`
	LocalVar       bool   `value:"false"`
	LocalGetter    bool   `value:"false"`
	LocalPrefix    string `value:"__"`
	InitMethod     string
	Init           []string
	typeName       string
}

func (s *Singleton) SetLocalVar(localVar bool) {
	s.LocalVar = localVar
	if localVar {
		s.LocalGetter = false
	}
}

func (s *Singleton) SetLocalGetter(localGetter bool) {
	s.LocalGetter = localGetter
	if localGetter {
		s.LocalVar = false
	}
}

func (s *Singleton) WriteString(buf io.StringWriter) error {
	if s.LocalVar {
		buf.WriteString(fmt.Sprintf("%s%s = ", s.LocalPrefix, s.typeName))
	}

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

	if s.LocalVar || !s.LazyInit {
		buf.WriteString(".Get()")
	} else if s.LocalGetter {
		buf.WriteString(".Getter()")
	}

	buf.WriteString("\n")

	return nil
}
