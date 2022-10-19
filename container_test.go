// nolint
package dimple

import (
	"context"
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

func TestNew(t *testing.T) {
	c := New(context.TODO())

	assert.NotNil(t, c)
}

func TestCircularDependency(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	ctn := New(context.TODO()).
		Add(Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceA := &randomService{
				Name: "A",
				B:    ctx.Container().Get(serviceB).(*randomService), // depends on b
			}

			return instanceA, nil
		}))).
		Add(Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceB := &randomService{
				Name: "B",
				C:    ctx.Container().Get(serviceC).(*randomService), // depends on c
			}

			return instanceB, nil
		}))).
		Add(Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceC := &randomService{
				Name: "C",
				A:    ctx.Container().Get(serviceA).(*randomService), // depends on A
			}

			return instanceC, nil
		})))

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
