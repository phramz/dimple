// nolint
package dimple

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParam(t *testing.T) {
	id := func(v any) string {
		return fmt.Sprintf(`%v`, v)
	}

	for _, tt := range []any{
		"string",
		1,
		0.9,
		map[string]string{"foo": "bar"},
		[]string{"a", "b", "c"},
		nil,
		struct {
			foo int
		}{foo: 1},
		make(chan any),
	} {
		ctn := Builder(Param(id(tt), tt)).MustBuild(context.TODO())
		assert.EqualValues(t, tt, ctn.MustGet(id(tt)))
	}
}
