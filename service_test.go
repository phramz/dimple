// nolint
package dimple

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceWithFn(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}
	ctn := New(context.TODO()).
		Add(Service(serviceA, WithFn(func() any {
			instanceA.Name = "A"

			return instanceA
		}))).
		Add(Service(serviceB, WithFn(func() any {
			instanceB.Name = "B"

			return instanceB
		}))).
		Add(Service(serviceC, WithFn(func() any {
			instanceC.Name = "C"

			return instanceC
		})))

	c := ctn.Get(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "C", c.Name)
	assert.Same(t, instanceC, c)

	b := ctn.Get(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "B", b.Name)
	assert.Same(t, instanceB, b)

	a := ctn.Get(serviceA).(*randomService)
	assert.NotNil(t, a)
	assert.EqualValues(t, "A", a.Name)
	assert.Same(t, instanceA, a)
}

func TestServiceWithError(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}
	ctn := New(context.TODO()).
		Add(Service(serviceA, WithErrorFn(func() (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		}))).
		Add(Service(serviceB, WithErrorFn(func() (any, error) {
			instanceB.Name = "B"

			return instanceB, nil
		}))).
		Add(Service(serviceC, WithErrorFn(func() (any, error) {
			instanceC.Name = "C"

			return instanceC, nil
		})))

	c := ctn.Get(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "C", c.Name)
	assert.Same(t, instanceC, c)

	b := ctn.Get(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "B", b.Name)
	assert.Same(t, instanceB, b)

	a := ctn.Get(serviceA).(*randomService)
	assert.NotNil(t, a)
	assert.EqualValues(t, "A", a.Name)
	assert.Same(t, instanceA, a)
}

func TestServiceWithContext(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}

	ctn := New(context.TODO()).
		Add(Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		}))).
		Add(Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			a := ctx.Container().Get(serviceA).(*randomService) // depends on a
			instanceB.Name = fmt.Sprintf(`%s->%s`, a.Name, "B")
			instanceB.A = a

			return instanceB, nil
		}))).
		Add(Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			b := ctx.Container().Get(serviceB).(*randomService) // depends on b
			instanceC.Name = fmt.Sprintf(`%s->%s`, b.Name, "C")
			instanceC.B = b

			return instanceC, nil
		})))

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
