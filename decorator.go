package dimple

var _ DecoratorDef = (*decoratorDef)(nil)

// Decorator returns a new instance of DecoratorDef
func Decorator(id string, decorates string, factory Factory) DecoratorDef {
	return &decoratorDef{
		definition: definition{
			id: id,
		},
		factory:   factory,
		decorates: decorates,
	}
}

type decoratorDef struct {
	definition
	factory   Factory
	instance  any
	decorates string
	decorated Definition
}

func (d *decoratorDef) Decorated() Definition {
	return d.decorated
}

func (d *decoratorDef) Factory() Factory {
	return d.factory
}

func (d *decoratorDef) Instance() any {
	return d.instance
}

func (d *decoratorDef) Decorates() string {
	return d.decorates
}

func (d *decoratorDef) WithID(id string) DecoratorDef {
	c := d.clone()
	c.id = id

	return c
}

func (d *decoratorDef) WithFactory(factory Factory) DecoratorDef {
	c := d.clone()
	c.factory = factory

	return c
}

func (d *decoratorDef) WithInstance(instance any) DecoratorDef {
	c := d.clone()
	c.instance = instance

	return c
}

func (d *decoratorDef) WithDecorates(id string) DecoratorDef {
	c := d.clone()
	c.decorates = id

	return c
}

func (d *decoratorDef) WithDecorated(def Definition) DecoratorDef {
	c := d.clone()
	c.decorated = def

	return c
}

func (d *decoratorDef) clone() *decoratorDef {
	return &decoratorDef{
		definition: *d.definition.clone(),
		factory:    d.Factory(),
		instance:   d.Instance(),
		decorates:  d.Decorates(),
		decorated:  d.Decorated(),
	}
}
