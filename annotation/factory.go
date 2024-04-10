package annotation

import (
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"go/ast"
	"strings"
)

// @Singleton
type Factory struct{}

func (f *Factory) Annotations() map[string][]api.AnnotationType {
	return map[string][]api.AnnotationType{
		AnnotationSingleton.Val(): {api.AnnotationTypeType},
	}
}

func (f *Factory) New(typedAnnotations []*api.TypedAnnotation) (api.Generator, error) {
	singletons := []*Singleton{}

	for _, ta := range typedAnnotations {
		if ta.Type == api.AnnotationTypeType {
			ts := ta.Node.(*ast.TypeSpec)

			for _, an := range ta.Annotations.Annotations {
				if strings.EqualFold(an.Name, AnnotationSingleton.Val()) {
					s := factory.New[Singleton]()
					err := an.To(s)
					if err != nil {
						return nil, err
					}

					s.typeName = ts.Name.Name

					singletons = append(singletons, s)
				}
			}
		}
	}

	if len(singletons) == 0 {
		return nil, nil
	}

	return newGenerator(singletons)
}

func (f *Factory) Order() api.Order {
	return api.OrderHigh
}
