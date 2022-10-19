package dimple

import "context"

// Container abstraction interface
type Container interface {
	// Add will add a newInstance Definition to the container. The val argument can be either a
	// - DecoratorDef if you want to decorate another service
	// - ServiceDef for a custom service definition
	// - ParamDef any parameter value of any type
	// Beware that With will panic when called after booting the Container explicit either (by calling Boot()) or
	// implicit by the first external Get call.
	Add(def Definition) Container

	// Get will return the value (for ParamDef) or the instance (for ServiceDef and DecoratorDef) by id
	// Beware that this can panic at runtime if any instantiation errors occur! Consider to explicitly call Boot()
	// before using it
	Get(id string) any

	// Has will return TRUE when a Definition by given id exist, otherwise false
	Has(id string) bool

	// GetDefinition returns a Definition by its ID, otherwise nil if it does not exist
	GetDefinition(id string) Definition

	// Boot will instantiate all service. It is not mandatory to call Boot() since all services (except decorated services)
	// will get instantiated on demand (lazy) per default.
	// If you want to ensure that every service could be instantiated it is recommended to call boot once all definitions
	// are set. Otherwise, any error throw during service instantiation will cause a panic at runtime.
	// Beware that after booting the Container it is no longer possible to call With()
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
}

// FactoryFn plain without anything
type FactoryFn = func() any

// FactoryFnWithError to define an anonymous functions
type FactoryFnWithError = func() (any, error)

// FactoryFnWithContext to define an anonymous functions
type FactoryFnWithContext = func(ctx FactoryCtx) (any, error)
