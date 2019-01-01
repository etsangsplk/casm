package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	id := New()
	assert.Equal(t, id, id.ID())
}
