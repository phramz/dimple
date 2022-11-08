package dimple

import (
	"context"
	"time"
)

var _ FactoryCtx = (*factoryContext)(nil)

func newFactoryCtx(ctx context.Context, c *DefaultContainer, d any) FactoryCtx {
	return &factoryContext{
		ctx:       ctx,
		container: c,
		decorated: d,
		serviceID: *c.ref,
	}
}

type factoryContext struct {
	ctx       context.Context
	container Container
	decorated any
	serviceID string
}

func (f *factoryContext) ServiceID() string {
	return f.serviceID
}

func (f *factoryContext) Ctx() context.Context {
	return f.ctx
}

func (f *factoryContext) Deadline() (deadline time.Time, ok bool) {
	return f.ctx.Deadline()
}

func (f *factoryContext) Done() <-chan struct{} {
	return f.ctx.Done()
}

func (f *factoryContext) Err() error {
	return f.ctx.Err()
}

func (f *factoryContext) Value(key any) any {
	return f.ctx.Value(key)
}

func (f *factoryContext) Container() Container {
	return f.container
}

func (f *factoryContext) Decorated() any {
	return f.decorated
}
