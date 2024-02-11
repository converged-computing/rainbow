package provider

import (
	"fmt"
	"sync/atomic"
)

var (
	_                     MessageIterator = (*CountedMessageProvider)(nil)
	MockedMessageProvider MessageIterator = &CountedMessageProvider{}
)

// MessageIterator is an interface that defines the behavior of a stream iterator.
type MessageIterator interface {
	Next() string
	HasNext() bool
}

// CountedMessageProvider is a simple message provider that generates a message with a counter.
type CountedMessageProvider struct {
	counter atomic.Uint64 // counter for messages
}

func (p *CountedMessageProvider) HasNext() bool {
	return true
}

func (p *CountedMessageProvider) Next() string {
	p.counter.Add(1)
	return fmt.Sprintf("message number %d", p.counter.Load())
}
