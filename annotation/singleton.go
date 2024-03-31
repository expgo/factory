package annotation

/*
	@Enum {
		Singleton
	}
*/
type Annotation string

type Singleton struct {
	Name           string
	NamedOnly      bool `value:"false"`
	UseConstructor bool `value:"false"`
	InitMethod     string
	Init           []string
	typeName       string
}
