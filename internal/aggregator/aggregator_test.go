package aggregator

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/stretchr/testify/assert"
)

// MockConfig is a mock implementation of the config.Config for testing.
type MockConfig struct {
	ConnectTwitch    bool
	ConnectYoutube   bool
	TwitchChannel    string
	YoutubeChannelId string
	YoutubeApiKey    string
}

// GetConfig is a mock implementation of the config.GetConfig for testing.
func (m *MockConfig) GetConfig() *config.Config {
	return &config.Config{
		ConnectTwitch:    m.ConnectTwitch,
		ConnectYoutube:   m.ConnectYoutube,
		TwitchChannel:    m.TwitchChannel,
		YoutubeChannelId: m.YoutubeChannelId,
		YoutubeApiKey:    m.YoutubeApiKey,
	}
}

// MockChatProvider is a mock implementation of the ChatProvider interface for testing.
type MockChatProvider struct {
	Name         string
	ConnectError error
	ListenError  error
	Messages     []chatmodels.ChatMessage
	Connected    bool
	Disconnected bool
	Config       *config.Config
}

func (m *MockChatProvider) Connect(cfx *config.Config) error {
	m.Config = cfx
	if m.ConnectError != nil {
		m.Connected = false
		return m.ConnectError
	}
	m.Connected = true
	return nil
}

func (m *MockChatProvider) Disconnect() error {
	m.Disconnected = true
	return nil
}

func (m *MockChatProvider) Listen(messages chan<- chatmodels.ChatMessage) error {
	if m.ListenError != nil {
		m.Connected = false
		return m.ListenError
	}
	var wg sync.WaitGroup

	// Create a channel to signal when all messages are sent
	done := make(chan struct{})

	// Send messages concurrently
	for _, msg := range m.Messages {
		wg.Add(1)
		go func(msg chatmodels.ChatMessage) {
			defer wg.Done()
			messages <- msg
		}(msg)
	}
	go func() { wg.Wait(); close(done) }() // Wait for all messages to be sent
	<-done
	return nil
}

func (m *MockChatProvider) GetName() string {
	return m.Name
}

// MockChatConsumer is a mock implementation of the ChatConsumer interface for testing.
type MockChatConsumer struct {
	Name     string
	Consumed []chatmodels.ChatMessage
	Stopped  bool
}

func (m *MockChatConsumer) Consume(message chatmodels.ChatMessage) {
	m.Consumed = append(m.Consumed, message)
}

func (m *MockChatConsumer) GetName() string {
	return m.Name
}

func TestAggregator_AddProvider(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	provider := &MockChatProvider{Name: "MockProvider"}
	agg.AddProvider(provider)

	assert.Equal(t, 1, agg.GetProvidersCount())
}

func TestAggregator_StartAndStop(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	provider := &MockChatProvider{Name: "MockProvider"}
	agg.AddProvider(provider)

	agg.Start()
	time.Sleep(100 * time.Millisecond) // Give some time for goroutines to start
	agg.Stop()

	assert.True(t, provider.Disconnected)
}

func TestAggregator_Start_NoProviders(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	agg.Start()
	agg.Stop()
}

func TestAggregator_ReceiveMessages(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	messages := []chatmodels.ChatMessage{
		{Provider: "MockProvider1", Content: "Message 1"},
		{Provider: "MockProvider1", Content: "Message 2"},
		{Provider: "MockProvider2", Content: "Message 3"},
	}

	provider1 := &MockChatProvider{Name: "MockProvider1", Messages: messages[:2]}
	provider2 := &MockChatProvider{Name: "MockProvider2", Messages: messages[2:]}
	agg.AddProvider(provider1)
	agg.AddProvider(provider2)

	agg.Start()
	defer agg.Stop()

	receivedMessages := make([]chatmodels.ChatMessage, 0)
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		for msg := range agg.GetMessages() {
			receivedMessages = append(receivedMessages, msg)
			if len(receivedMessages) == len(messages) {
				done <- true
				return
			}
		}
	}()

	select {
	case <-done:
		assert.ElementsMatch(t, messages, receivedMessages)

	case <-timeout:
		t.Fatalf("Timeout waiting for messages. Received: %v", receivedMessages)
	}
}

func TestAggregator_ProviderConnectError(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	provider := &MockChatProvider{Name: "MockProvider", ConnectError: fmt.Errorf("connect error")}
	agg.AddProvider(provider)

	agg.Start()
	defer agg.Stop()

	// Because of the connection error the chat provider should not remain connected
	assert.False(t, provider.Connected)
}

func TestAggregator_ProviderListenError(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	provider := &MockChatProvider{Name: "MockProvider", ListenError: fmt.Errorf("listen error")}
	agg.AddProvider(provider)

	agg.Start()
	defer agg.Stop()

	// Because of the error while listening the chat provider should not remain connected
	assert.False(t, provider.Connected)
}

func TestAggregator_ConsumeMessages(t *testing.T) {
	mockCfg := &MockConfig{}
	cfg := mockCfg.GetConfig()
	agg := NewAggregator(cfg)

	messages := []chatmodels.ChatMessage{
		{Provider: "MockProvider1", Content: "Message 1"},
		{Provider: "MockProvider1", Content: "Message 2"},
		{Provider: "MockProvider2", Content: "Message 3"},
	}

	provider1 := &MockChatProvider{Name: "MockProvider1", Messages: messages[:2]}
	provider2 := &MockChatProvider{Name: "MockProvider2", Messages: messages[2:]}
	agg.AddProvider(provider1)
	agg.AddProvider(provider2)

	consumer1 := &MockChatConsumer{Name: "MockConsumer1"}
	consumer2 := &MockChatConsumer{Name: "MockConsumer2"}
	agg.AddConsumer(consumer1)
	agg.AddConsumer(consumer2)

	agg.Start()
	defer agg.Stop()

	// Wait for a short time to allow messages to be processed.
	time.Sleep(500 * time.Millisecond)

	assert.Len(t, consumer1.Consumed, len(messages))
	assert.Len(t, consumer2.Consumed, len(messages))
	assert.ElementsMatch(t, messages, consumer1.Consumed)
	assert.ElementsMatch(t, messages, consumer2.Consumed)
}
