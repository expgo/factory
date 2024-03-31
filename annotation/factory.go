package annotation

import (
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"go/ast"
)

func init() {
	factory.Singleton[Factory]()
}

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
			if an := ta.Annotations.FindAnnotationByName(AnnotationSingleton.Val()); an != nil {
				s := &Singleton{}
				err := an.To(s)
				if err != nil {
					return nil, err
				}

				ts := ta.Node.(*ast.TypeSpec)
				s.typeName = ts.Name.Name

				singletons = append(singletons, s)
			}
		}
	}

	if len(singletons) == 0 {
		return nil, nil
	}

	return newGenerator(singletons)
}
