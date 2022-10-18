package pkg

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/thoas/go-funk"
)

var _ Container = (*container)(nil)

// New returns a new container instance
func New(ctx context.Context) Container {
	return &container{
		ctx:         ctx,
		order:       make([]string, 0),
		definitions: make(map[string]Definition),
	}
}

type container struct {
	sync.Mutex
	booted      bool
	order       []string
	ref         *string
	parent      *container
	ctx         context.Context
	definitions map[string]Definition
}

func (c *container) Ctx() context.Context {
	return c.ctx
}

func (c *container) Boot() error {
	if err := c.boot(c.allDecoratorIDs()...); err != nil {
		return err
	}

	return c.boot(c.allServiceIDs()...)
}

func (c *container) Definition(id string) Definition {

	if c.parent != nil {
		return c.parent.Definition(id)
	}

	if def, ok := c.definitions[id]; ok {
		return def
	}

	return nil
}

func (c *container) Has(id string) bool {
	return c.Definition(id) != nil
}

func (c *container) Get(id string) any {
	if !c.booted {
		if err := c.boot(c.allDecoratorIDs()...); err != nil {
			panic(err)
		}
	}

	instance, err := c.get(id)
	if err != nil {
		panic(err)
	}

	return instance
}

func (c *container) With(id string, val any) Container {
	if c.booted {
		panic(ErrContainerAlreadyBooted)
	}

	return c.with(id, val)
}

func (c *container) with(id string, val any) Container {
	c.Lock()
	defer c.Unlock()

	var def Definition
	switch t := val.(type) {
	case DecoratorDef:
		def = t.WithID(id)
	case ServiceDef:
		def = t.WithID(id)
	case ParamDef:
		def = t.WithID(id)
	case Definition:
		panic(fmt.Sprintf(`unsupported type of definiton "%T" for service "%s"`, val, id))
	case FactoryFn:
		def = Service(t).WithID(id)
	default:
		def = Param(val).WithID(id)
	}

	if c.parent == nil {
		c.definitions[id] = def
		c.order = append(c.order, id)
		return c
	}

	c.parent.with(id, def)

	return c
}

func (c *container) get(id string) (any, error) {
	var targetInstance any
	var instance any
	var err error

	def := c.Definition(id)
	if def == nil {
		return nil, fmt.Errorf(`%w: cannot find definiton for service "%s"`, ErrUnknownService, id)
	}

	// if it's a ParamDef just return the value
	if svc, ok := def.(ParamDef); ok {
		return svc.Value(), nil
	}

	if c.isCircularDependency(id) {
		return nil, fmt.Errorf(`%w: %s`, ErrCircularDependency, c.debugPathInfo(c.path(id)))
	}

	indirection := c.indirect(id)
	if svc, ok := def.(ServiceDef); ok {
		if svc.Instance() != nil {
			// already instantiated?
			return svc.Instance(), nil
		}

		// we need to instantiate a new service
		instance, err = svc.Fn()(newFactoryCtx(indirection.ctx, indirection, targetInstance))
		if err != nil {
			return nil, fmt.Errorf(`%w: cannot instantiate service "%s: %s"`, ErrServiceFactoryFailed, id, err.Error())
		}

		c.with(id, svc.WithInstance(instance))

		return instance, nil
	}

	if svc, ok := def.(DecoratorDef); ok {
		if svc.Instance() != nil {
			// already instantiated?
			return svc.Instance(), nil
		}

		// we need to instantiate a new service
		if indirection.isCircularDependency(svc.Decorates()) {
			return nil, fmt.Errorf(`%w: %s`, ErrCircularDependency, indirection.debugPathInfo(indirection.path(svc.Decorates())))
		}

		targetDef := indirection.Definition(svc.Decorates())
		targetInstance, err = indirection.get(svc.Decorates())
		if err != nil {
			return nil, err
		}

		instance, err = svc.Fn()(newFactoryCtx(indirection.ctx, indirection, targetInstance))
		if err != nil {
			return nil, fmt.Errorf(`%w: cannot instantiate service "%s: %s"`, ErrServiceFactoryFailed, id, err.Error())
		}

		target := svc.WithInstance(instance).
			WithDecorated(targetDef)

		c.with(id, target)
		c.with(svc.Decorates(), target)

		return instance, nil
	}

	panic(fmt.Sprintf(`unsupported type of definiton "%T" for service "%s"`, def, id))
}

func (c *container) indirect(id string) *container {
	indirection := c.clone()
	indirection.ref = &id

	return indirection
}

func (c *container) isCircularDependency(id string) bool {
	if c.ref != nil && *c.ref == id {
		return true
	}

	if c.parent == nil {
		return false
	}

	return c.parent.isCircularDependency(id)
}

func (c *container) path(id string) []string {
	path := make([]string, 0)
	if c.ref != nil {
		path = append(path, *c.ref)
	}

	path = append(path, id)

	indirection := c.parent
	for indirection != nil && indirection.ref != nil {
		path = append([]string{*indirection.ref}, path...)

		indirection = indirection.parent
	}

	return path
}

func (c *container) boot(ids ...string) error {
	if c.parent != nil {
		return nil
	}

	defer func() {
		c.booted = true
	}()

	for _, id := range c.order {
		if !funk.ContainsString(ids, id) {
			continue
		}

		if _, err := c.get(id); err != nil {
			return err
		}
	}

	return nil
}

func (c *container) allDecoratorIDs() []string {
	return funk.FilterString(funk.Keys(c.definitions).([]string), func(s string) bool {
		def := c.definitions[s]
		_, ok := def.(DecoratorDef)

		return ok
	})
}

func (c *container) allServiceIDs() []string {
	return funk.FilterString(funk.Keys(c.definitions).([]string), func(s string) bool {
		def := c.definitions[s]
		_, ok := def.(ServiceDef)

		return ok
	})
}

func (c *container) clone() *container {
	return &container{
		ctx:         c.ctx,
		definitions: c.definitions,
		parent:      c,
	}
}

func (c *container) debugPathInfo(defIDs []string) string {
	pathInfo := make([]string, 0)
	for _, defID := range defIDs {
		subPath := c.debugDecoratorPath(c.Definition(defID))
		subPathInfo := ""

		if len(subPath) > 0 {
			subPathInfo = fmt.Sprintf(`:decorates(%s)`, strings.Join(subPath, ``))
		}

		pathInfo = append(pathInfo, fmt.Sprintf(`"%s"%s`, defID, subPathInfo))
	}

	return strings.Join(pathInfo, ` -> `)
}

func (c *container) debugDecoratorPath(def Definition) []string {
	path := make([]string, 0)
	if def == nil {
		return path
	}

	dec, ok := def.(DecoratorDef)
	if !ok {
		return path
	}

	path = append(path, fmt.Sprintf(`"%s"`, dec.Decorates()))

	if dec.Decorated() != nil {
		path = append(path, c.debugDecoratorPath(dec.Decorated())...)
	}

	return path
}
