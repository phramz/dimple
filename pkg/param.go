package pkg

var _ ParamDef = (*paramDef)(nil)

// Param returns a new regular value definition
func Param(v any) ParamDef {
	return &paramDef{
		definition: definition{},
		value:      v,
	}
}

type paramDef struct {
	definition
	value any
}

func (p *paramDef) Value() any {
	return p.value
}

func (p *paramDef) clone() *paramDef {
	return &paramDef{
		definition: *p.definition.clone(),
		value:      p.Value(),
	}
}

func (p *paramDef) WithID(id string) ParamDef {
	c := p.clone()
	c.id = id

	return c
}

func (p *paramDef) WithValue(v any) ParamDef {
	c := p.clone()
	c.value = v

	return c
}
