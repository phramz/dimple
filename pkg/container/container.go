package container

import (
	"fmt"
	"strings"
	"sync"

	"github.com/phramz/dimple/pkg/contract"
)

var _ contract.Container = (*container)(nil)

func New() contract.Container {
	return &container{
		factories: make(map[string]contract.FactoryFn),
		instances: make(map[string]interface{}),
	}
}

type container struct {
	sync.Mutex
	factories map[string]contract.FactoryFn
	instances map[string]interface{}
	parent    *container
	ref       *string
}

func (c *container) clone() *container {
	return &container{
		factories: c.factories,
		instances: c.instances,
		parent:    c,
	}
}

func (c *container) getFactory(id string) contract.FactoryFn {
	c.Lock()
	defer c.Unlock()

	if services, ok := c.factories[id]; ok {
		return services
	}

	if c.parent == nil {
		return nil
	}

	return c.parent.getFactory(id)
}

func (c *container) getInstance(id string) interface{} {
	c.Lock()
	defer c.Unlock()

	if instance, ok := c.instances[id]; ok {
		return instance
	}

	if c.parent == nil {
		return nil
	}

	return c.parent.getInstance(id)
}

func (c *container) setInstance(id string, instance interface{}) interface{} {
	c.Lock()
	defer c.Unlock()

	c.instances[id] = instance
	if c.parent != nil {
		c.instances[id] = c.parent.setInstance(id, instance)
	}

	return c.instances[id]
}

func (c *container) Get(id string) interface{} {
	if instance := c.getInstance(id); instance != nil {
		return instance
	}

	factory := c.getFactory(id)
	if factory == nil {
		panic(fmt.Errorf(`%w: cannot instantiate service "%s"`, contract.ErrUnknownService, id))
	}

	if c.isCircularDependency(id) {
		panic(fmt.Errorf(`%w: %s`, contract.ErrCircularDependency, strings.Join(c.getPath(id), ` -> `)))
	}

	clone := c.clone()
	clone.ref = &id

	instance, err := factory(clone)
	if err != nil {
		panic(fmt.Errorf(`%w: cannot instantiate service "%s: %s"`, contract.ErrServiceFactoryFailed, id, err.Error()))
	}

	return clone.setInstance(id, instance)
}

func (c *container) With(id string, val interface{}) contract.Container {
	c.Lock()
	defer c.Unlock()

	fn, ok := val.(contract.FactoryFn)
	if !ok {
		fn = func(container contract.Container) (interface{}, error) {
			return val, nil
		}
	}

	c.factories[id] = fn

	if c.parent != nil {
		c.parent.With(id, fn)
	}

	return c
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

func (c *container) getPath(id string) []string {
	path := make([]string, 0)
	if c.ref != nil {
		path = append(path, *c.ref)
	}

	path = append(path, id)

	lvl := c.parent
	for lvl != nil && lvl.ref != nil {
		path = append([]string{*lvl.ref}, path...)

		lvl = lvl.parent
	}

	return path
}
