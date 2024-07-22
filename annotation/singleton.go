package annotation

import (
	"fmt"
	"io"
	"strings"
)

type Singleton struct {
	Name            string
	NamedOnly       bool   `value:"false"`
	UseConstructor  bool   `value:"false"`
	LocalGetter     bool   `value:"false"`
	LocalPrefix     string `value:"__"`
	LocalGetterName string
	InitMethod      string
	Init            []string
	typeName        string
}

func (s *Singleton) WriteString(buf io.StringWriter) error {
	if s.LocalGetter {
		localGetterName := ""
		if len(s.LocalGetterName) > 0 {
			localGetterName = s.LocalGetterName
		} else if len(s.Name) > 0 {
			localGetterName = fmt.Sprintf("%s%s_%s", s.LocalPrefix, s.typeName, s.Name)
		} else {
			localGetterName = fmt.Sprintf("%s%s", s.LocalPrefix, s.typeName)
		}

		buf.WriteString(fmt.Sprintf("%s = factory.Getter[%s](", localGetterName, s.typeName))
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

	if s.LocalGetter {
		buf.WriteString(")")
	}

	buf.WriteString("\n")

	return nil
}
