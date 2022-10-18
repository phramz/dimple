package dimple

var _ ServiceDef = (*serviceDef)(nil)

// Service returns a new regular service definition
func Service(fn FactoryFn) ServiceDef {
	return &serviceDef{
		definition: definition{},
		fn:         fn,
	}
}

type serviceDef struct {
	definition
	fn       FactoryFn
	instance any
}

func (s *serviceDef) clone() *serviceDef {
	return &serviceDef{
		definition: *s.definition.clone(),
		fn:         s.Fn(),
		instance:   s.Instance(),
	}
}

func (s *serviceDef) WithID(id string) ServiceDef {
	c := s.clone()
	c.id = id

	return c
}

func (s *serviceDef) WithFn(fn FactoryFn) ServiceDef {
	c := s.clone()
	c.fn = fn

	return c
}

func (s *serviceDef) WithInstance(instance any) ServiceDef {
	c := s.clone()
	c.instance = instance

	return c
}

func (s *serviceDef) Instance() any {
	return s.instance
}

func (s *serviceDef) Fn() FactoryFn {
	return s.fn
}
