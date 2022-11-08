package dimple

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBuilder_MustBuild(t *testing.T) {
	b := Builder()
	assert.NotNil(t, b)
	assert.Implements(t, (*ContainerBuilder)(nil), b)

	c := b.MustBuild(context.TODO())
	assert.NotNil(t, c)
	assert.Implements(t, (*Container)(nil), c)
}

func TestDefaultBuilder_Add(t *testing.T) {
	def1 := Param("def1", &struct{}{})
	def2 := Param("def2", &struct{}{})

	b := Builder(def1)
	assert.True(t, b.Has("def1"))
	assert.Same(t, def1.Value(), b.Get("def1").(ParamDef).Value())

	b.Add(def2)
	assert.True(t, b.Has("def2"))
	assert.Same(t, def2.Value(), b.Get("def2").(ParamDef).Value())
}
