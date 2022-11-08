package dimple

import (
	"context"
)

var _ ContainerBuilder = (*DefaultBuilder)(nil)

// Builder returns a new ContainerBuilder instance
func Builder(defs ...Definition) *DefaultBuilder {
	b := &DefaultBuilder{
		container: &DefaultContainer{
			order:       make([]string, 0),
			definitions: make(map[string]Definition),
		},
	}

	b.Add(Service("container", WithInstance(b.container)))

	for _, def := range defs {
		b.Add(def)
	}

	return b
}

type DefaultBuilder struct {
	container *DefaultContainer
}

func (b *DefaultBuilder) MustBuild(ctx context.Context) *DefaultContainer {
	c, err := b.Build(ctx)
	if err != nil {
		panic(err)
	}

	return c
}

func (b *DefaultBuilder) Build(ctx context.Context) (*DefaultContainer, error) {
	c := b.container
	c.ctx = ctx

	b.Add(Service("context", WithInstance(ctx)))

	// mandatory boot of decorated services to rewrite the decorated definitions
	if err := c.boot(c.getAllDecoratorIDs()...); err != nil {
		return nil, err
	}

	return c, nil
}

func (b *DefaultBuilder) Add(def Definition) ContainerBuilder {
	b.container.add(def.Id(), def)

	return b
}

func (b *DefaultBuilder) Get(id string) Definition {
	return b.container.getDefinition(id)
}

func (b *DefaultBuilder) Has(id string) bool {
	return b.container.Has(id)
}
