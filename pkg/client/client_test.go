package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	t.Run("new client sans address", func(t *testing.T) {
		_, err := NewClient("")
		assert.Error(t, err)
	})

	c := &SimpleClient{}

	t.Run("close client", func(t *testing.T) {
		err := c.Close()
		assert.NoError(t, err)
		assert.Empty(t, c.GetTarget())
	})

	t.Run("serial sans message", func(t *testing.T) {
		_, err := c.Serial(context.TODO(), "")
		assert.Error(t, err)
	})

	t.Run("stream sans iterator", func(t *testing.T) {
		err := c.Stream(context.TODO(), nil)
		assert.Error(t, err)
	})
}
