// nolint
package pkg

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type randomInterface interface {
	SayMyName() string
}

type randomService struct {
	A    *randomService
	B    *randomService
	C    *randomService
	Name string
}

func (r *randomService) SayMyName() string {
	return r.Name
}

type decoratorService struct {
	randomService
	Decorated randomInterface
}

func (d *decoratorService) SayMyName() string {
	return fmt.Sprintf(`%s%s`, d.Decorated.SayMyName(), d.Name)
}

func TestNew(t *testing.T) {
	c := New(context.TODO())

	assert.NotNil(t, c)
}

func TestWithDecorator(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	ctn := New(context.TODO()).
		With(serviceA, func(ctx FactoryCtx) (any, error) {
			instanceA := &randomService{
				Name: "A",
			}

			return instanceA, nil
		}).
		With(serviceB, Decorator(serviceA, func(ctx FactoryCtx) (any, error) {
			instanceB := &decoratorService{
				randomService: randomService{
					Name: "B",
				},
				Decorated: ctx.Decorated().(randomInterface),
			}

			return instanceB, nil
		})).
		With(serviceC, Decorator(serviceA, func(ctx FactoryCtx) (any, error) {
			instanceC := &decoratorService{
				randomService: randomService{
					Name: "C",
				},
				Decorated: ctx.Decorated().(randomInterface),
			}

			return instanceC, nil
		}))

	c := ctn.Get(serviceC)
	assert.NotNil(t, c)
	assert.IsType(t, &decoratorService{}, c)
	assert.EqualValues(t, "ABC", c.(randomInterface).SayMyName())

	b := ctn.Get(serviceB)
	assert.NotNil(t, b)
	assert.IsType(t, &decoratorService{}, b)
	assert.EqualValues(t, "AB", b.(randomInterface).SayMyName())

	a := ctn.Get(serviceA)
	assert.NotNil(t, a)
	assert.IsType(t, &decoratorService{}, a)
	assert.EqualValues(t, "ABC", a.(randomInterface).SayMyName())
}

func TestWithFactoryFn(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}
	ctn := New(context.TODO()).
		With(serviceA, func(ctx FactoryCtx) (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		}).
		With(serviceB, func(ctx FactoryCtx) (any, error) {
			a := ctx.Container().Get(serviceA).(*randomService) // depends on a
			instanceB.Name = fmt.Sprintf(`%s->%s`, a.Name, "B")
			instanceB.A = a

			return instanceB, nil
		}).
		With(serviceC, func(ctx FactoryCtx) (any, error) {
			b := ctx.Container().Get(serviceB).(*randomService) // depends on b
			instanceC.Name = fmt.Sprintf(`%s->%s`, b.Name, "C")
			instanceC.B = b

			return instanceC, nil
		})

	c := ctn.Get(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "A->B->C", c.Name)
	assert.Nil(t, c.A)
	assert.NotNil(t, c.B)
	assert.Same(t, instanceB, c.B)
	assert.NotNil(t, c.B.A)
	assert.Same(t, instanceA, c.B.A)

	b := ctn.Get(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "A->B", b.Name)
	assert.Same(t, instanceB, b)
	assert.NotNil(t, b.A)
	assert.Same(t, instanceA, b.A)
	assert.Nil(t, b.B)
	assert.Nil(t, b.C)

	a := ctn.Get(serviceA).(*randomService)
	assert.NotNil(t, a)
	assert.EqualValues(t, "A", a.Name)
	assert.Same(t, instanceA, a)
	assert.Nil(t, a.A)
	assert.Nil(t, a.B)
	assert.Nil(t, a.C)
}

func TestWithServiceDef(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}

	ctn := New(context.TODO()).
		With(serviceA, Service(func(ctx FactoryCtx) (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		})).
		With(serviceB, Service(func(ctx FactoryCtx) (any, error) {
			a := ctx.Container().Get(serviceA).(*randomService) // depends on a
			instanceB.Name = fmt.Sprintf(`%s->%s`, a.Name, "B")
			instanceB.A = a

			return instanceB, nil
		})).
		With(serviceC, Service(func(ctx FactoryCtx) (any, error) {
			b := ctx.Container().Get(serviceB).(*randomService) // depends on b
			instanceC.Name = fmt.Sprintf(`%s->%s`, b.Name, "C")
			instanceC.B = b

			return instanceC, nil
		}))

	c := ctn.Get(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "A->B->C", c.Name)
	assert.Nil(t, c.A)
	assert.NotNil(t, c.B)
	assert.Same(t, instanceB, c.B)
	assert.NotNil(t, c.B.A)
	assert.Same(t, instanceA, c.B.A)

	b := ctn.Get(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "A->B", b.Name)
	assert.Same(t, instanceB, b)
	assert.NotNil(t, b.A)
	assert.Same(t, instanceA, b.A)
	assert.Nil(t, b.B)
	assert.Nil(t, b.C)

	a := ctn.Get(serviceA).(*randomService)
	assert.NotNil(t, a)
	assert.EqualValues(t, "A", a.Name)
	assert.Same(t, instanceA, a)
	assert.Nil(t, a.A)
	assert.Nil(t, a.B)
	assert.Nil(t, a.C)
}

func TestCircularDependency(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	ctn := New(context.TODO()).
		With(serviceA, func(ctx FactoryCtx) (any, error) {
			instanceA := &randomService{
				Name: "A",
				B:    ctx.Container().Get(serviceB).(*randomService), // depends on b
			}

			return instanceA, nil
		}).
		With(serviceB, func(ctx FactoryCtx) (any, error) {
			instanceB := &randomService{
				Name: "B",
				C:    ctx.Container().Get(serviceC).(*randomService), // depends on c
			}

			return instanceB, nil
		}).
		With(serviceC, func(ctx FactoryCtx) (any, error) {
			instanceC := &randomService{
				Name: "C",
				A:    ctx.Container().Get(serviceA).(*randomService), // depends on A
			}

			return instanceC, nil
		})

	defer func() {
		if r := recover(); r != nil {
			assert.ErrorIs(t, r.(error), ErrCircularDependency)
			assert.Contains(t, r.(error).Error(), `"service.a" -> "service.b" -> "service.c" -> "service.a"`)
			return
		}

		t.Errorf("Expected a panic!")
	}()

	_ = ctn.Get(serviceA).(*randomService)
}
