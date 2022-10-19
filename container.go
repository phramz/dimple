package dimple

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/thoas/go-funk"
)

var _ Container = (*container)(nil)

// New returns a newInstance container instance
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

func (c *container) Inject(target any) error {
	if !c.isInjectable(target) {
		return fmt.Errorf(`unable to inject to target. it has to been a pointer to a struct an addressable, got "%T"`, target)
	}

	v := reflect.ValueOf(target).Elem()
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		if id, ok := typeField.Tag.Lookup("dimple"); ok {
			instance := c.Get(id)
			fieldVal := v.Field(i)
			if !fieldVal.CanSet() {
				return fmt.Errorf(`unable to inject value to field "%s" since it is not writable`, typeField.Name)
			}

			fieldVal.Set(reflect.ValueOf(instance))
		}
	}

	return nil
}

func (c *container) isInjectable(target any) bool {
	if reflect.ValueOf(target).Kind() != reflect.Pointer {
		return false
	}

	if reflect.ValueOf(target).Elem().Kind() != reflect.Struct {
		return false
	}

	if !reflect.ValueOf(target).Elem().CanAddr() {
		return false
	}

	return true
}

func (c *container) Add(def Definition) Container {
	if c.booted {
		panic(ErrContainerAlreadyBooted)
	}

	return c.add(def.Id(), def)
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

func (c *container) GetDefinition(id string) Definition {

	if c.parent != nil {
		return c.parent.GetDefinition(id)
	}

	if def, ok := c.definitions[id]; ok {
		return def
	}

	return nil
}

func (c *container) Has(id string) bool {
	return c.GetDefinition(id) != nil
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

func (c *container) add(id string, val any) Container {
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
	default:
		panic(fmt.Sprintf(`unsupported type definition %T`, val))
	}

	if c.parent == nil {
		c.definitions[id] = def
		c.order = append(c.order, id)
		return c
	}

	c.parent.add(id, def)

	return c
}

func (c *container) newInstance(def Definition) (any, error) {
	var err error
	var target, instance any
	var f Factory

	if svc, ok := def.(ServiceDef); ok {
		f = svc.Factory()
	}

	if svc, ok := def.(DecoratorDef); ok {
		f = svc.Factory()
	}

	if f == nil {
		return nil, fmt.Errorf(`%w: cannot instantiate service "%s" due to missing factory`, ErrServiceFactoryFailed, def.Id())
	}

	if svc, ok := def.(DecoratorDef); ok {
		target, err = c.get(svc.Decorates())
		if err != nil {
			return nil, err
		}
	}

	if instance = f.Instance(); instance != nil {
		return instance, nil
	}

	if fn := f.FactoryFnWithContext(); fn != nil {
		instance, err = fn(newFactoryCtx(c.ctx, c, target))
		if err != nil {
			return nil, fmt.Errorf(`%w: cannot instantiate service "%s: %s"`, ErrServiceFactoryFailed, def.Id(), err.Error())
		}

		return instance, err
	}

	if fn := f.FactoryFnWithError(); fn != nil {
		instance, err = fn()
		if err != nil {
			return nil, fmt.Errorf(`%w: cannot instantiate service "%s: %s"`, ErrServiceFactoryFailed, def.Id(), err.Error())
		}

		return instance, err
	}

	if fn := f.FactoryFn(); fn != nil {
		instance = fn()
		if instance == nil {
			return nil, fmt.Errorf(`%w: cannot instantiate service "%s: factory returned nil"`, ErrServiceFactoryFailed, def.Id())
		}

		return instance, nil
	}

	return nil, fmt.Errorf(`%w: cannot instantiate service "%s: no factory function provided"`, ErrServiceFactoryFailed, def.Id())
}

func (c *container) get(id string) (any, error) {
	def := c.GetDefinition(id)
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
		return indirection.getService(svc)
	}

	if svc, ok := def.(DecoratorDef); ok {
		return indirection.getDecoration(svc)
	}

	panic(fmt.Sprintf(`unsupported type of definiton "%T" for service "%s"`, def, id))
}

func (c *container) getDecoration(svc DecoratorDef) (any, error) {
	if svc.Instance() != nil {
		// already instantiated?
		return svc.Instance(), nil
	}

	// we need to instantiate a newInstance service
	if c.isCircularDependency(svc.Decorates()) {
		return nil, fmt.Errorf(`%w: %s`, ErrCircularDependency, c.debugPathInfo(c.path(svc.Decorates())))
	}

	instance, err := c.newInstance(svc)
	if err != nil {
		return nil, err
	}

	if c.isInjectable(instance) {
		if err = c.Inject(instance); err != nil {
			return nil, err
		}
	}

	targetDef := c.GetDefinition(svc.Decorates())
	if err != nil {
		return nil, err
	}

	target := svc.WithInstance(instance).WithDecorated(targetDef)
	c.add(svc.Id(), target)
	c.add(svc.Decorates(), target)

	return instance, nil
}

func (c *container) getService(def ServiceDef) (any, error) {
	if def.Instance() != nil {
		// already instantiated?
		return def.Instance(), nil
	}

	// we need to instantiate a newInstance service
	instance, err := c.newInstance(def)
	if err != nil {
		return nil, err
	}

	if c.isInjectable(instance) {
		if err = c.Inject(instance); err != nil {
			return nil, err
		}
	}

	c.add(def.Id(), def.WithInstance(instance))

	return instance, nil
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
		subPath := c.debugDecoratorPath(c.GetDefinition(defID))
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
