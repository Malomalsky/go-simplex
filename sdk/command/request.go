package command

type Request interface {
	CommandString() string
}

type Raw string

func (r Raw) CommandString() string {
	return string(r)
}

func Lookup(name string) (Definition, bool) {
	for _, def := range GeneratedCatalog {
		if def.Name == name {
			return def, true
		}
	}
	return Definition{}, false
}

