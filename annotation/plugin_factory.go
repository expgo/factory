package annotation

import (
	"fmt"
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"go/ast"
	"strings"
)

// @Singleton
type PluginFactory struct{}

func (f *PluginFactory) Annotations() map[string][]api.AnnotationType {
	return map[string][]api.AnnotationType{
		AnnotationSingleton.Val(): {api.AnnotationTypeType},
		AnnotationFactory.Val():   {api.AnnotationTypeFunc},
	}
}

func (f *PluginFactory) New(typedAnnotations []*api.TypedAnnotation) (api.Generator, error) {
	singletons := []*Singleton{}
	factories := []*Factory{}

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

		if ta.Type == api.AnnotationTypeFunc {
			fd := ta.Node.(*ast.FuncDecl)

			for _, an := range ta.Annotations.Annotations {
				if strings.EqualFold(an.Name, AnnotationFactory.Val()) {
					fac := factory.New[Factory]()
					err := an.To(fac)
					if err != nil {
						return nil, err
					}

					fac.funcName = fd.Name.Name
					if fd.Type.Results.NumFields() != 1 {
						return nil, fmt.Errorf("%s's return only be one", fac.funcName)
					}

					fac.funcReturn = fd.Type.Results.List[0].Type.(*ast.Ident).Name

					factories = append(factories, fac)
				}
			}
		}
	}

	if len(singletons) == 0 && len(factories) == 0 {
		return nil, nil
	}

	return newGenerator(singletons, factories)
}

func (f *PluginFactory) Order() api.Order {
	return api.OrderHigh
}
