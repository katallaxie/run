package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVars_Merge(t *testing.T) {
	m := make(Vars)
	m.Merge(Vars{"foo": "bar"})

	assert.Equal(t, Vars{"foo": "bar"}, m)

	m.Merge(Vars{"foo": "bar2"})

	assert.Equal(t, Vars{"foo": "bar2"}, m)
}
