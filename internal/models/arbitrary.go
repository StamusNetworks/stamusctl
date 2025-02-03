package models

import "strconv"

type Arbitrary map[string]any

func NewArbitrary() Arbitrary {
	return Arbitrary{}
}

func (a *Arbitrary) SetArbitrary(arbitrary map[string]string) {
	for key, value := range arbitrary {
		(*a)[key] = asLooseTyped(value)
	}
}

func (a *Arbitrary) AsMap() map[string]any {
	return *a
}

func (a *Arbitrary) Set(key string, value any) {
	(*a)[key] = value
}

// Return the value of a string to any type
func asLooseTyped(value string) any {
	if value == "true" || value == "false" {
		return value == "true"
	}
	asInt, err := strconv.Atoi(value)
	if err != nil {
		return value
	}
	return asInt
}
