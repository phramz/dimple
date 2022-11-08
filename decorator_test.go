// nolint
package dimple

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type decoratorService struct {
	randomService
	Decorated randomInterface
}

func (d *decoratorService) SayMyName() string {
	return fmt.Sprintf(`%s%s`, d.Decorated.SayMyName(), d.Name)
}

func TestDecoratorWithFn(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	builder := Builder(
		Service(serviceA, WithFn(func() any {
			instanceA := &randomService{
				Name: "A",
			}

			return instanceA
		})),
		Decorator(serviceB, serviceA, WithFn(func() any {
			instanceB := &decoratorService{
				randomService: randomService{
					Name: "B",
				},
			}

			return instanceB
		})),
		Decorator(serviceC, serviceA, WithFn(func() any {
			instanceC := &decoratorService{
				randomService: randomService{
					Name: "C",
				},
			}

			return instanceC
		})),
	)

	container := builder.MustBuild(context.TODO())

	a := container.MustGet(serviceA)
	assert.NotNil(t, a)
	assert.IsType(t, &decoratorService{}, a)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))

	b := container.MustGet(serviceB)
	assert.NotNil(t, b)
	assert.IsType(t, &decoratorService{}, b)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))

	c := container.MustGet(serviceC)
	assert.NotNil(t, c)
	assert.IsType(t, &decoratorService{}, c)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))
}

func TestDecoratorWithContext(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	builder := Builder(
		Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceA := &randomService{
				Name: "A",
			}

			return instanceA, nil
		})),
		Decorator(serviceB, serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceB := &decoratorService{
				randomService: randomService{
					Name: "B",
				},
				Decorated: ctx.Decorated().(randomInterface),
			}

			return instanceB, nil
		})),
		Decorator(serviceC, serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceC := &decoratorService{
				randomService: randomService{
					Name: "C",
				},
				Decorated: ctx.Decorated().(randomInterface),
			}

			return instanceC, nil
		})),
	)

	container := builder.MustBuild(context.TODO())
	c := container.MustGet(serviceC)
	assert.NotNil(t, c)
	assert.IsType(t, &decoratorService{}, c)
	assert.EqualValues(t, "ABC", c.(randomInterface).SayMyName())

	b := container.MustGet(serviceB)
	assert.NotNil(t, b)
	assert.IsType(t, &decoratorService{}, b)
	assert.EqualValues(t, "AB", b.(randomInterface).SayMyName())

	a := container.MustGet(serviceA)
	assert.NotNil(t, a)
	assert.IsType(t, &decoratorService{}, a)
	assert.EqualValues(t, "ABC", a.(randomInterface).SayMyName())
}

func TestDecoratorWithError(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	builder := Builder(
		Service(serviceA, WithErrorFn(func() (any, error) {
			instanceA := &randomService{
				Name: "A",
			}

			return instanceA, nil
		})),
		Decorator(serviceB, serviceA, WithErrorFn(func() (any, error) {
			instanceB := &decoratorService{
				randomService: randomService{
					Name: "B",
				},
			}

			return instanceB, nil
		})),
		Decorator(serviceC, serviceA, WithErrorFn(func() (any, error) {
			instanceC := &decoratorService{
				randomService: randomService{
					Name: "C",
				},
			}

			return instanceC, nil
		})),
	)

	container := builder.MustBuild(context.TODO())

	a := container.MustGet(serviceA)
	assert.NotNil(t, a)
	assert.IsType(t, &decoratorService{}, a)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))

	b := container.MustGet(serviceB)
	assert.NotNil(t, b)
	assert.IsType(t, &decoratorService{}, b)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))

	c := container.MustGet(serviceC)
	assert.NotNil(t, c)
	assert.IsType(t, &decoratorService{}, c)
	assert.IsType(t, &decoratorDef{}, builder.Get(serviceA))
}
