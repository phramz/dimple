package dimple

var _ ServiceDef = (*serviceDef)(nil)

// Service returns a newInstance regular service definition
func Service(id string, factory Factory) ServiceDef {
	return &serviceDef{
		definition: definition{
			id: id,
		},
		factory: factory,
	}
}

type serviceDef struct {
	definition
	factory  Factory
	instance any
}

func (s *serviceDef) clone() *serviceDef {
	return &serviceDef{
		definition: *s.definition.clone(),
		factory:    s.Factory(),
		instance:   s.Instance(),
	}
}

func (s *serviceDef) WithID(id string) ServiceDef {
	c := s.clone()
	c.id = id

	return c
}

func (s *serviceDef) WithFactory(factory Factory) ServiceDef {
	c := s.clone()
	c.factory = factory

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

func (s *serviceDef) Factory() Factory {
	return s.factory
}
