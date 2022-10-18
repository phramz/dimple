package pkg

import "context"

// Container abstraction interface
type Container interface {
	// With will add a new Definition to the container. The val argument can be either a
	// - FactoryFn for a plain service definition
	// - DecoratorDef if you want to decorate another service
	// - ServiceDef for a custom service definition
	// - ParamDef any parameter value of any type
	// ... any other type is implicitly considered a ParamDef
	// Beware that With will panic when called after booting the Container explicit either (by calling Boot()) or
	// implicit by the first external Get call.
	With(id string, val any) Container

	// Get will return the value (for ParamDef) or the instance (for ServiceDef and DecoratorDef) by id
	// Beware that this can panic at runtime if any instantiation errors occur! Consider to explicitly call Boot()
	// before using it
	Get(id string) any

	// Has will return TRUE when a Definition by given id exist, otherwise false
	Has(id string) bool

	// Definition returns a Definition by its ID, otherwise nil if it does not exist
	Definition(id string) Definition

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
	Fn() FactoryFn
	Instance() any
	WithID(id string) ServiceDef
	WithFn(fn FactoryFn) ServiceDef
	WithInstance(instance any) ServiceDef
}

// DecoratorDef abstraction interface
type DecoratorDef interface {
	Definition
	Fn() FactoryFn
	Instance() any
	Decorates() string
	Decorated() Definition
	WithID(id string) DecoratorDef
	WithFn(fn FactoryFn) DecoratorDef
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

// FactoryFn to define factory functions
type FactoryFn = func(ctx FactoryCtx) (any, error)
