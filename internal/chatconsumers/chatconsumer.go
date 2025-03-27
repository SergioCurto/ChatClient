package chatconsumers

import (
	"fmt"

	"github.com/SergioCurto/ChatClient/internal/chatconsumers/console"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
)

type ChatConsumer interface {
	Consume(message chatmodels.ChatMessage)
	GetName() string
}

type ChatConsumerType int

const (
	Console ChatConsumerType = iota
)

// ChatConsumerFactory is the factory interface for creating ChatConsumers.
type ChatConsumerFactory interface {
	CreateConsumer(consumerType ChatConsumerType) (ChatConsumer, error)
}

// ConcreteChatConsumerFactory is a concrete implementation of the ChatConsumerFactory.
type ConcreteChatConsumerFactory struct {
}

// NewConcreteChatConsumerFactory creates a new ConcreteChatConsumerFactory.
func NewConcreteChatConsumerFactory() *ConcreteChatConsumerFactory {
	return &ConcreteChatConsumerFactory{}
}

// CreateConsumer creates a ChatConsumer based on the given consumerType.
func (f *ConcreteChatConsumerFactory) CreateConsumer(consumerType ChatConsumerType) (ChatConsumer, error) {
	switch consumerType {
	case Console:
		return &console.ConsoleConsumer{Name: "Console"}, nil
	default:
		return nil, fmt.Errorf("unknown consumer type: %v", consumerType)
	}
}
