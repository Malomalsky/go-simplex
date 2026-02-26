package command

type Request interface {
	CommandString() string
}

type ResponseAwareRequest interface {
	Request
	ExpectedResponseTypes() []string
}

type Raw string

func (r Raw) CommandString() string {
	return string(r)
}

func ExpectedResponseTypes(req Request) []string {
	if req == nil {
		return nil
	}
	typed, ok := req.(ResponseAwareRequest)
	if !ok {
		return nil
	}
	return typed.ExpectedResponseTypes()
}

func Lookup(name string) (Definition, bool) {
	for _, def := range GeneratedCatalog {
		if def.Name == name {
			return def, true
		}
	}
	return Definition{}, false
}
