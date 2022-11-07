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

type injectableService struct {
	InjectedA randomInterface `inject:"service.a"`
	InjectedB randomInterface `inject:"service.b"`
	InjectedC randomInterface `inject:"service.c"`
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

func TestContainer_Inject(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"
	const serviceD = "service.d"

	out := &injectableService{}
	ctn := New(context.TODO()).
		Add(Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "A"}, nil
		}))).
		Add(Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "B"}, nil
		}))).
		Add(Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "C"}, nil
		}))).
		Add(Service(serviceD, WithInstance(out)))

	actual := ctn.Get(serviceD).(*injectableService)
	assert.Same(t, out, actual)

	assert.Same(t, ctn.Get(serviceA), actual.InjectedA)
	assert.Same(t, ctn.Get(serviceB), actual.InjectedB)
	assert.Same(t, ctn.Get(serviceC), actual.InjectedC)

	assert.Equal(t, "A", actual.InjectedA.SayMyName())
	assert.Equal(t, "B", actual.InjectedB.SayMyName())
	assert.Equal(t, "C", actual.InjectedC.SayMyName())
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
