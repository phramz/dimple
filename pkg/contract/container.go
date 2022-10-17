package contract

type Container interface {
	With(id string, val interface{}) Container
	Get(id string) interface{}
}

type FactoryFn = func(c Container) (interface{}, error)
