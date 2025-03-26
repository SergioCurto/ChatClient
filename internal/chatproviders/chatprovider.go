package chatproviders

import (
	"fmt"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/SergioCurto/ChatClient/internal/chatproviders/twitch"
	"github.com/SergioCurto/ChatClient/internal/chatproviders/youtube"
)

type ChatProvider interface {
	Connect(cfx *config.Config) error
	Disconnect() error
	Listen(messages chan<- chatmodels.ChatMessage) error
	GetName() string
}

type ChatProviderType int

const (
	Twitch ChatProviderType = iota
	Youtube
)

// ChatProviderFactory is the factory interface for creating ChatProviders.
type ChatProviderFactory interface {
	CreateProvider(providerType string) (ChatProvider, error)
}

// ConcreteChatProviderFactory is a concrete implementation of the ChatProviderFactory.
type ConcreteChatProviderFactory struct {
}

// NewConcreteChatProviderFactory creates a new ConcreteChatProviderFactory.
func NewConcreteChatProviderFactory() *ConcreteChatProviderFactory {
	return &ConcreteChatProviderFactory{}
}

// CreateProvider creates a ChatProvider based on the given providerType.
func (f *ConcreteChatProviderFactory) CreateProvider(providerType ChatProviderType) (ChatProvider, error) {
	switch providerType {
	case Twitch:
		return twitch.NewTwitchProvider(), nil
	case Youtube:
		return youtube.NewYoutubeProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %v", providerType)
	}
}
