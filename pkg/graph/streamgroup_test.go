package graph

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type multiFail struct {
	sync.Once
	closed     bool
	err0, err1 error
}

func (mf *multiFail) Close() error {
	var first bool
	mf.closed = true
	if mf.Do(func() { first = true }); first {
		return mf.err0
	}
	return mf.err1
}

func TestIdempotentCloser(t *testing.T) {
	ic := &idempotentCloser{
		Closer: &multiFail{err0: nil, err1: errors.New("not idempotent")},
	}
	assert.NoError(t, ic.Close())
	assert.NoError(t, ic.Close())
}

func TestMultiCloser(t *testing.T) {
	t.Run("IsIdempotent", func(t *testing.T) {
		err0 := &multiFail{err0: nil, err1: errors.New("not idempotent")}
		err1 := &multiFail{err0: nil, err1: errors.New("not idempotent")}
		err2 := &multiFail{err0: nil, err1: errors.New("not idempotent")}

		mc := newMultiCloser(err0, err1, err2)
		assert.NoError(t, mc.Close())
		assert.NoError(t, mc.Close())

		for i, err := range []*multiFail{err0, err1, err2} {
			assert.True(t, err.closed, fmt.Sprintf("number %d not closed", i))
		}
	})
}

func TestStreamGroup(t *testing.T) {
	t.Run("TopDownClose", func(t *testing.T) {
		sg := newStreamGroup(newStream(nil), newStream(nil))
		assert.NoError(t, sg.Close())

		assert.Error(t, sg.Context().Err())
		assert.Error(t, sg.data.Context().Err())
		assert.Error(t, sg.ctrl.Context().Err())
	})

	t.Run("BottomUpClose", func(t *testing.T) {
		t.Run("Data", func(t *testing.T) {
			sg := newStreamGroup(newStream(nil), newStream(nil))
			assert.NoError(t, sg.data.Close())

			assert.Error(t, sg.Context().Err())
			assert.Error(t, sg.data.Context().Err())
			assert.Error(t, sg.ctrl.Context().Err())
		})

		t.Run("Ctrl", func(t *testing.T) {
			sg := newStreamGroup(newStream(nil), newStream(nil))
			assert.NoError(t, sg.ctrl.Close())

			assert.Error(t, sg.Context().Err())
			assert.Error(t, sg.data.Context().Err())
			assert.Error(t, sg.ctrl.Context().Err())
		})
	})
}
