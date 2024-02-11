package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageIterator(t *testing.T) {
	it := &CountedMessageProvider{}
	assert.NotNil(t, it)

	// Test the iterator
	p := ""
	c := 0
	for it.HasNext() {
		m := it.Next()
		assert.NotEmpty(t, m)
		assert.NotEqual(t, p, m)
		p = m
		c++
		if c > 10 {
			break
		}
	}
}
