package console

import (
	"fmt"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
)

// ConsoleConsumer is a ChatConsumer that logs messages to the console.
type ConsoleConsumer struct {
	Name string
}

func NewConsoleConsumer() *ConsoleConsumer {
	return &ConsoleConsumer{
		Name: "Console",
	}
}

func (c *ConsoleConsumer) Start(cfg *config.Config) error {
	return nil
}

// Consume logs the message to the console.
func (c *ConsoleConsumer) Consume(message chatmodels.ChatMessage) {
	fmt.Printf("[%s] %s\n", message.Provider, message.Content)
}

// GetName returns the name of the consumer.
func (c *ConsoleConsumer) GetName() string {
	return c.Name
}
