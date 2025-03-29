package chatconsumers

import (
	"fmt"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatconsumers/console"
	"github.com/SergioCurto/ChatClient/internal/chatconsumers/simplepage"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
)

type ChatConsumer interface {
	Consume(message chatmodels.ChatMessage)
	Start(cfg *config.Config) error
	GetName() string
}

type ChatConsumerType int

const (
	Console ChatConsumerType = iota
	SimplePage
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
		return console.NewConsoleConsumer(), nil
	case SimplePage:
		return simplepage.NewSimplePageConsumer(), nil
	default:
		return nil, fmt.Errorf("unknown consumer type: %v", consumerType)
	}
}
