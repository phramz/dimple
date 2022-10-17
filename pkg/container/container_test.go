package container

import (
	"fmt"
	"github.com/phramz/dimple/pkg/contract"
	"github.com/stretchr/testify/assert"
	"testing"
)

type randomService struct {
	Name    string
	A, B, C *randomService
}

func TestNew(t *testing.T) {
	c := New()

	assert.NotNil(t, c)
}

func TestService(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	instanceA := &randomService{}
	instanceB := &randomService{}
	instanceC := &randomService{}

	ctn := New().
		With(serviceA, func(c contract.Container) (interface{}, error) {
			instanceA.Name = "A"

			return instanceA, nil
		}).
		With(serviceB, func(c contract.Container) (interface{}, error) {
			a := c.Get(serviceA).(*randomService) // depends on a
			instanceB.Name = fmt.Sprintf(`%s->%s`, a.Name, "B")
			instanceB.A = a

			return instanceB, nil
		}).
		With(serviceC, func(c contract.Container) (interface{}, error) {
			b := c.Get(serviceB).(*randomService) // depends on b
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

func TestCircularDependency(t *testing.T) {
	const serviceA = "service.a"
	const serviceB = "service.b"
	const serviceC = "service.c"

	ctn := New().
		With(serviceA, func(c contract.Container) (interface{}, error) {
			instanceA := &randomService{
				Name: "A",
				B:    c.Get(serviceB).(*randomService), // depends on b
			}

			return instanceA, nil
		}).
		With(serviceB, func(c contract.Container) (interface{}, error) {
			instanceB := &randomService{
				Name: "B",
				C:    c.Get(serviceC).(*randomService), // depends on c
			}

			return instanceB, nil
		}).
		With(serviceC, func(c contract.Container) (interface{}, error) {
			instanceC := &randomService{
				Name: "C",
				A:    c.Get(serviceA).(*randomService), // depends on A
			}

			return instanceC, nil
		})

	defer func() {
		if r := recover(); r != nil {
			assert.ErrorIs(t, r.(error), contract.ErrCircularDependency)
			assert.Contains(t, r.(error).Error(), `service.a -> service.b -> service.c -> service.a`)
			return
		}

		t.Errorf("Expected a panic!")
	}()

	_ = ctn.Get(serviceA).(*randomService)
}
