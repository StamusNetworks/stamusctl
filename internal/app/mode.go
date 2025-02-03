package app

type ModeStruct string

func (m *ModeStruct) set(value string) {
	*m = ModeStruct(value)
}

func (m *ModeStruct) IsTest() bool {
	return *m == "test"
}

func (m *ModeStruct) IsProd() bool {
	return *m == "prod"
}

type EmbedStruct string

func (e *EmbedStruct) Set(value string) {
	*e = EmbedStruct(value)
}

func (e *EmbedStruct) IsTrue() bool {
	return *e == "true"
}
