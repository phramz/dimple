package pkg

var _ DecoratorDef = (*decoratorDef)(nil)

// Decorator returns a new regular service definition
func Decorator(decorates string, fn FactoryFn) DecoratorDef {
	return &decoratorDef{
		definition: definition{},
		fn:         fn,
		decorates:  decorates,
	}
}

type decoratorDef struct {
	definition
	fn        FactoryFn
	instance  any
	decorates string
	decorated Definition
}

func (d *decoratorDef) Decorated() Definition {
	return d.decorated
}

func (d *decoratorDef) Fn() FactoryFn {
	return d.fn
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

func (d *decoratorDef) WithFn(fn FactoryFn) DecoratorDef {
	c := d.clone()
	c.fn = fn

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
		fn:         d.Fn(),
		instance:   d.Instance(),
		decorates:  d.Decorates(),
		decorated:  d.Decorated(),
	}
}
