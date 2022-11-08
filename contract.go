package dimple

import "context"

// ContainerBuilder abstraction interface
type ContainerBuilder interface {
	// Add will add a new Definition to the ContainerBuilder. The val argument can be either a
	// - ServiceDef for a service
	// - DecoratorDef if you want to decorate another service
	// - ParamDef any parameter value of any type
	Add(def Definition) ContainerBuilder

	// Get returns a Definition by its ID, otherwise nil if it does not exist
	Get(id string) Definition

	// Has returns TRUE if a Definition of given ID exists
	Has(id string) bool
}

// Container abstraction interface
type Container interface {
	// Has will return TRUE when a service or param by given id exist, otherwise FALSE
	Has(id string) bool

	// Get will return a plain value (for ParamDef) or the instance (for ServiceDef and DecoratorDef) by id
	Get(id string) (any, error)

	// MustGet will return the param value or service instance by id
	// Beware that this can panic at runtime if any instantiation errors occur!
	// Consider to explicitly call Boot() before using it
	MustGet(id string) any

	// Inject will take a struct as target and sets the field values according to tagged service ids.
	// Example:
	//
	// type MyStruct struct {
	//     TimeService     *TimeService   `inject:"service.time"`
	//     TimeFormat      string         `inject:"param.time_format"`
	// }
	Inject(target any) error

	// Boot will instantiate all services eagerly. It is not mandatory to call Boot() since all
	// services (except decorated services) will be instantiated lazy per default.
	// Beware that lazy instantiation can cause a panic at runtime when retrieving values via MustGet()!
	// If you want to ensure that every service can be instantiated properly it is recommended to call Boot()
	// before first use of MustGet().
	Boot() error

	// Ctx returns the context.Context
	Ctx() context.Context
}

// Definition abstraction interface
type Definition interface {
	Id() string
}

// ParamDef abstraction interface
type ParamDef interface {
	Definition
	Value() any
	WithID(id string) ParamDef
	WithValue(v any) ParamDef
}

// ServiceDef abstraction interface
type ServiceDef interface {
	Definition
	Factory() Factory
	Instance() any
	WithID(id string) ServiceDef
	WithFactory(factory Factory) ServiceDef
	WithInstance(instance any) ServiceDef
}

// DecoratorDef abstraction interface
type DecoratorDef interface {
	Definition
	Factory() Factory
	Instance() any
	Decorates() string
	Decorated() Definition
	WithID(id string) DecoratorDef
	WithFactory(factory Factory) DecoratorDef
	WithInstance(instance any) DecoratorDef
	WithDecorates(id string) DecoratorDef
	WithDecorated(def Definition) DecoratorDef
}

type FactoryCtx interface {
	context.Context
	Ctx() context.Context
	Container() Container
	Decorated() any
	ServiceID() string
}

type Factory interface {
	FactoryFn() FactoryFn
	FactoryFnWithError() FactoryFnWithError
	FactoryFnWithContext() FactoryFnWithContext
	Instance() any
}

// FactoryFn plain without anything
type FactoryFn = func() any

// FactoryFnWithError to define an anonymous functions
type FactoryFnWithError = func() (any, error)

// FactoryFnWithContext to define an anonymous functions
type FactoryFnWithContext = func(ctx FactoryCtx) (any, error)
