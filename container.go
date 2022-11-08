package dimple

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/thoas/go-funk"
)

var _ Container = (*DefaultContainer)(nil)

type DefaultContainer struct {
	sync.Mutex
	booted      bool
	order       []string
	ref         *string
	parent      *DefaultContainer
	ctx         context.Context
	definitions map[string]Definition
}

// MustGetT generic wrapper for Container.MustGet
func MustGetT[T any](c Container, id string) T {
	val := c.MustGet(id)
	valT, ok := val.(T)
	if !ok {
		panic(fmt.Sprintf(`illegal type assertion for service "%s" of type "%T"`, id, val))
	}

	return valT
}

func (c *DefaultContainer) Inject(target any) error {
	if !c.isInjectable(target) {
		return fmt.Errorf(`unable to inject to target. it has to been a pointer to a struct an addressable, got "%T"`, target)
	}

	v := reflect.ValueOf(target).Elem()
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		if id, ok := typeField.Tag.Lookup("inject"); ok {
			instance, err := c.Get(id)
			if err != nil {
				return err
			}

			fieldVal := v.Field(i)
			if !fieldVal.CanSet() {
				return fmt.Errorf(`unable to inject value to field "%s" since it is not writable`, typeField.Name)
			}

			fieldVal.Set(reflect.ValueOf(instance))
		}
	}

	return nil
}

func (c *DefaultContainer) Boot() error {
	// decorated services must be instantiated first
	if err := c.boot(c.getAllDecoratorIDs()...); err != nil {
		return err
	}

	return c.boot(c.getAllServiceIDs()...)
}

func (c *DefaultContainer) Has(id string) bool {
	return c.getDefinition(id) != nil
}

func (c *DefaultContainer) MustGet(id string) any {
	instance, err := c.Get(id)
	if err != nil {
		panic(err)
	}

	return instance
}

func (c *DefaultContainer) Get(id string) (any, error) {
	if !c.booted {
		if err := c.boot(c.getAllDecoratorIDs()...); err != nil {
			panic(err)
		}
	}

	instance, err := c.getValue(id)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (c *DefaultContainer) Ctx() context.Context {
	return c.ctx
}

func (c *DefaultContainer) boot(ids ...string) error {
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

		if _, err := c.getValue(id); err != nil {
			return err
		}
	}

	return nil
}

func (c *DefaultContainer) add(id string, val any) *DefaultContainer {
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
		panic(fmt.Sprintf(`unsupported type getDefinition %T`, val))
	}

	if c.parent == nil {
		c.definitions[id] = def
		c.order = append(c.order, id)
		return c
	}

	c.parent.add(id, def)

	return c
}

func (c *DefaultContainer) getInstance(def Definition) (any, error) {
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
		target, err = c.getValue(svc.Decorates())
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

func (c *DefaultContainer) getValue(id string) (any, error) {
	if !c.Has(id) {
		return nil, fmt.Errorf(`%w: cannot find definiton for service "%s"`, ErrUnknownService, id)
	}

	def := c.getDefinition(id)

	// if it's a ParamDef just return the value
	if svc, ok := def.(ParamDef); ok {
		return svc.Value(), nil
	}

	if c.isCircularDependency(id) {
		return nil, fmt.Errorf(`%w: %s`, ErrCircularDependency, c.getDebugPathInfo(c.getPath(id)))
	}

	indirection := c.getIndirect(id)
	if svc, ok := def.(ServiceDef); ok {
		return indirection.getService(svc)
	}

	if svc, ok := def.(DecoratorDef); ok {
		return indirection.getDecoration(svc)
	}

	panic(fmt.Sprintf(`unsupported type of definiton "%T" for service "%s"`, def, id))
}

func (c *DefaultContainer) getDefinition(id string) Definition {

	if c.parent != nil {
		return c.parent.getDefinition(id)
	}

	if def, ok := c.definitions[id]; ok {
		return def
	}

	return nil
}

func (c *DefaultContainer) getDecoration(svc DecoratorDef) (any, error) {
	if svc.Instance() != nil {
		// already instantiated?
		return svc.Instance(), nil
	}

	// we need to instantiate a getInstance service
	if c.isCircularDependency(svc.Decorates()) {
		return nil, fmt.Errorf(`%w: %s`, ErrCircularDependency, c.getDebugPathInfo(c.getPath(svc.Decorates())))
	}

	instance, err := c.getInstance(svc)
	if err != nil {
		return nil, err
	}

	if c.isInjectable(instance) {
		if err = c.Inject(instance); err != nil {
			return nil, err
		}
	}

	if !c.Has(svc.Decorates()) {
		return nil, fmt.Errorf(`%w: cannot decorate non existinmg service "%s"`, ErrUnknownService, svc.Decorates())
	}

	targetDef := c.getDefinition(svc.Decorates())
	if err != nil {
		return nil, err
	}

	target := svc.WithInstance(instance).WithDecorated(targetDef)
	c.add(svc.Id(), target)
	c.add(svc.Decorates(), target)

	return instance, nil
}

func (c *DefaultContainer) getService(def ServiceDef) (any, error) {
	if def.Instance() != nil {
		// already instantiated?
		return def.Instance(), nil
	}

	// we need to instantiate a getInstance service
	instance, err := c.getInstance(def)
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

func (c *DefaultContainer) getIndirect(id string) *DefaultContainer {
	indirection := c.clone()
	indirection.ref = &id

	return indirection
}

func (c *DefaultContainer) isCircularDependency(id string) bool {
	if c.ref != nil && *c.ref == id {
		return true
	}

	if c.parent == nil {
		return false
	}

	return c.parent.isCircularDependency(id)
}

func (c *DefaultContainer) getPath(id string) []string {
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

func (c *DefaultContainer) getAllDecoratorIDs() []string {
	return funk.FilterString(funk.Keys(c.definitions).([]string), func(s string) bool {
		def := c.definitions[s]
		_, ok := def.(DecoratorDef)

		return ok
	})
}

func (c *DefaultContainer) getAllServiceIDs() []string {
	return funk.FilterString(funk.Keys(c.definitions).([]string), func(s string) bool {
		def := c.definitions[s]
		_, ok := def.(ServiceDef)

		return ok
	})
}

func (c *DefaultContainer) isInjectable(target any) bool {
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

func (c *DefaultContainer) getDebugPathInfo(defIDs []string) string {
	pathInfo := make([]string, 0)
	for _, defID := range defIDs {
		subPath := c.getDebugDecoratorPath(c.getDefinition(defID))
		subPathInfo := ""

		if len(subPath) > 0 {
			subPathInfo = fmt.Sprintf(`:decorates(%s)`, strings.Join(subPath, ``))
		}

		pathInfo = append(pathInfo, fmt.Sprintf(`"%s"%s`, defID, subPathInfo))
	}

	return strings.Join(pathInfo, ` -> `)
}

func (c *DefaultContainer) getDebugDecoratorPath(def Definition) []string {
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
		path = append(path, c.getDebugDecoratorPath(dec.Decorated())...)
	}

	return path
}

func (c *DefaultContainer) clone() *DefaultContainer {
	return &DefaultContainer{
		ctx:         c.ctx,
		definitions: c.definitions,
		parent:      c,
	}
}
