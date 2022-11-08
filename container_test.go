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

func TestContainer_Inject(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"
	const serviceD = "service.d"

	out := &injectableService{}
	ctn := Builder(
		Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "A"}, nil
		})),
		Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "B"}, nil
		})),
		Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			return &randomService{Name: "C"}, nil
		})),
		Service(serviceD, WithInstance(out)),
	).
		MustBuild(context.TODO())

	actual := ctn.MustGet(serviceD).(*injectableService)
	assert.Same(t, out, actual)

	assert.Same(t, ctn.MustGet(serviceA), actual.InjectedA)
	assert.Same(t, ctn.MustGet(serviceB), actual.InjectedB)
	assert.Same(t, ctn.MustGet(serviceC), actual.InjectedC)

	assert.Equal(t, "A", actual.InjectedA.SayMyName())
	assert.Equal(t, "B", actual.InjectedB.SayMyName())
	assert.Equal(t, "C", actual.InjectedC.SayMyName())
}

func TestCircularDependency(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	ctn := Builder(
		Service(serviceA, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceA := &randomService{
				Name: "A",
				B:    ctx.Container().MustGet(serviceB).(*randomService), // depends on b
			}

			return instanceA, nil
		})),
		Service(serviceB, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceB := &randomService{
				Name: "B",
				C:    ctx.Container().MustGet(serviceC).(*randomService), // depends on c
			}

			return instanceB, nil
		})),
		Service(serviceC, WithContextFn(func(ctx FactoryCtx) (any, error) {
			instanceC := &randomService{
				Name: "C",
				A:    ctx.Container().MustGet(serviceA).(*randomService), // depends on A
			}

			return instanceC, nil
		})),
	).MustBuild(context.TODO())

	defer func() {
		if r := recover(); r != nil {
			assert.ErrorIs(t, r.(error), ErrCircularDependency)
			assert.Contains(t, r.(error).Error(), `"service.a" -> "service.b" -> "service.c" -> "service.a"`)
			return
		}

		t.Errorf("Expected a panic!")
	}()

	_ = ctn.MustGet(serviceA).(*randomService)
}
