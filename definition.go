package dimple

var _ Definition = (*definition)(nil)

type definition struct {
	id string
}

func (s *definition) clone() *definition {
	return &definition{
		id: s.id,
	}
}

func (s *definition) Id() string {
	return s.id
}
