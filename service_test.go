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
	builder := Builder(
		Service(serviceA, WithFn(func() any {
			instanceA.Name = "A"

			return instanceA
		})),
		Service(serviceB, WithFn(func() any {
			instanceB.Name = "B"

			return instanceB
		})),
		Service(serviceC, WithFn(func() any {
			instanceC.Name = "C"

			return instanceC
		})),
	)

	container := builder.MustBuild(context.TODO())

	c := container.MustGet(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "C", c.Name)
	assert.Same(t, instanceC, c)

	b := container.MustGet(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "B", b.Name)
	assert.Same(t, instanceB, b)

	a := container.MustGet(serviceA).(*randomService)
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
	builder := Builder(
		Service(serviceA, WithErrorFn(func() (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		})),
		Service(serviceB, WithErrorFn(func() (any, error) {
			instanceB.Name = "B"

			return instanceB, nil
		})),
		Service(serviceC, WithErrorFn(func() (any, error) {
			instanceC.Name = "C"

			return instanceC, nil
		})),
	)

	container := builder.MustBuild(context.TODO())

	c := container.MustGet(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "C", c.Name)
	assert.Same(t, instanceC, c)

	b := container.MustGet(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "B", b.Name)
	assert.Same(t, instanceB, b)

	a := container.MustGet(serviceA).(*randomService)
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

	builder := Builder(
		Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceA.Name = "A"

			return instanceA, nil
		})),
		Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			a := ctx.Container().MustGet(serviceA).(*randomService) // depends on a
			instanceB.Name = fmt.Sprintf(`%s->%s`, a.Name, "B")
			instanceB.A = a

			return instanceB, nil
		})),
		Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			b := ctx.Container().MustGet(serviceB).(*randomService) // depends on b
			instanceC.Name = fmt.Sprintf(`%s->%s`, b.Name, "C")
			instanceC.B = b

			return instanceC, nil
		})),
	)

	container := builder.MustBuild(context.TODO())

	c := container.MustGet(serviceC).(*randomService)
	assert.NotNil(t, c)
	assert.EqualValues(t, "A->B->C", c.Name)
	assert.Nil(t, c.A)
	assert.NotNil(t, c.B)
	assert.Same(t, instanceB, c.B)
	assert.NotNil(t, c.B.A)
	assert.Same(t, instanceA, c.B.A)

	b := container.MustGet(serviceB).(*randomService)
	assert.NotNil(t, b)
	assert.EqualValues(t, "A->B", b.Name)
	assert.Same(t, instanceB, b)
	assert.NotNil(t, b.A)
	assert.Same(t, instanceA, b.A)
	assert.Nil(t, b.B)
	assert.Nil(t, b.C)

	a := container.MustGet(serviceA).(*randomService)
	assert.NotNil(t, a)
	assert.EqualValues(t, "A", a.Name)
	assert.Same(t, instanceA, a)
	assert.Nil(t, a.A)
	assert.Nil(t, a.B)
	assert.Nil(t, a.C)
}
