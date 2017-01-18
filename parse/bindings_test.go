package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindings(t *testing.T) {
	bindingsMap, err := Bindings([]byte(`
a:
  labels:
    a: b
  ports:
  - 8000
b:
  labels:
  - c=d
  ports:
  - 9000:9000`))
	assert.Nil(t, err)

	assert.Equal(t, bindingsMap["a"].Labels, map[string]string{
		"a": "b",
	})
	assert.Equal(t, bindingsMap["a"].Ports, []string{
		"8000",
	})
	assert.Equal(t, bindingsMap["b"].Labels, map[string]string{
		"c": "d",
	})
	assert.Equal(t, bindingsMap["b"].Ports, []string{
		"9000:9000",
	})
}
