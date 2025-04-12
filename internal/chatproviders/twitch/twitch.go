package twitch

import (
	"fmt"
	"log"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchProvider struct {
	// Twitch specific variables
	Name         string
	ShortName    string
	client       *twitch.Client
	channel      string
	messagesChan chan<- chatmodels.ChatMessage
}

func NewTwitchProvider() *TwitchProvider {
	return &TwitchProvider{
		Name:      "Twitch",
		ShortName: "Tw",
	}
}

func (t *TwitchProvider) Connect(cfx *config.Config) error {
	fmt.Println("Connecting to Twitch...")

	// Get Twitch credentials from environment variables
	t.channel = cfx.TwitchChannel
	if t.channel == "" {
		return fmt.Errorf("missing twitch_channel in environment variables")
	}

	// Create a new Twitch client
	t.client = twitch.NewAnonymousClient()

	// Join the specified channel
	t.client.Join(t.channel)

	return nil
}

func (t *TwitchProvider) Disconnect() error {
	fmt.Println("Disconnecting from Twitch...")
	if t.client != nil {
		t.client.Disconnect()
	}
	return nil
}

func (t *TwitchProvider) Listen(messages chan<- chatmodels.ChatMessage) error {
	t.messagesChan = messages

	// Handle incoming messages
	t.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		t.messagesChan <- chatmodels.ChatMessage{
			Provider:          t.GetName(),
			ProviderShortName: t.GetShortName(),
			Timestamp:         message.Time,
			Content:           message.Message,
			AuthorName:        message.User.DisplayName,
		}
	})

	// Handle connection errors
	t.client.OnConnect(func() {
		log.Println("Connected to Twitch chat")
	})

	// Start listening for messages
	go func() {
		err := t.client.Connect()
		if err != nil {
			log.Println("Error connecting to Twitch:", err)
		}
	}()

	return nil
}

func (t *TwitchProvider) GetName() string {
	return t.Name
}

func (t *TwitchProvider) GetShortName() string {
	return t.ShortName
}
