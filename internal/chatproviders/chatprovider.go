package chatproviders

import (
	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
)

type ChatProvider interface {
	Connect(cfx *config.Config) error
	Disconnect() error
	Listen(messages chan<- chatmodels.ChatMessage) error
	GetName() string
}
