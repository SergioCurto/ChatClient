package aggregator

import (
	"errors"
	"testing"
	"time"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/stretchr/testify/assert"
)

// MockChatProvider for testing
type MockChatProvider struct {
	Name          string
	ConnectErr    error
	DisconnectErr error
	ListenErr     error
	Messages      []chatmodels.ChatMessage
	Connected     bool
	Disconnected  bool
	ListenCalled  bool
	stopListen    chan struct{}
}

func (m *MockChatProvider) Connect(cfx *config.Config) error {
	m.Connected = true
	return m.ConnectErr
}

func (m *MockChatProvider) Disconnect() error {
	m.Disconnected = true
	close(m.stopListen)
	return m.DisconnectErr
}

func (m *MockChatProvider) Listen(messages chan<- chatmodels.ChatMessage) error {
	m.stopListen = make(chan struct{})
	m.ListenCalled = true
	go func() {
		for _, msg := range m.Messages {
			select {
			case messages <- msg:
			case <-m.stopListen:
				return
			}
		}
	}()
	return m.ListenErr
}

func (m *MockChatProvider) GetName() string {
	return m.Name
}

// MockChatConsumer for testing
type MockChatConsumer struct {
	Name             string
	ConsumedCount    int
	ConsumedMessages []chatmodels.ChatMessage
	StartCalled      bool
}

func (m *MockChatConsumer) Start(cfg *config.Config) error {
	m.StartCalled = true
	return nil
}

func (m *MockChatConsumer) Consume(message chatmodels.ChatMessage) {
	m.ConsumedCount++
	m.ConsumedMessages = append(m.ConsumedMessages, message)
}

func (m *MockChatConsumer) GetName() string {
	return m.Name
}

func TestNewAggregator(t *testing.T) {
	cfg := &config.Config{}
	agg := NewAggregator(cfg)
	assert.NotNil(t, agg)
	assert.NotNil(t, agg.messages)
	assert.Equal(t, cfg, agg.cfg)
	assert.Empty(t, agg.providers)
	assert.Empty(t, agg.consumers)
	assert.Nil(t, agg.stop)
}

func TestAggregator_AddProvider(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	provider := &MockChatProvider{Name: "TestProvider"}
	agg.AddProvider(provider)
	assert.Len(t, agg.providers, 1)
	assert.Equal(t, provider, agg.providers[0])
}

func TestAggregator_AddConsumer(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	consumer := &MockChatConsumer{Name: "TestConsumer"}
	agg.AddConsumer(consumer)
	assert.Len(t, agg.consumers, 1)
	assert.Equal(t, consumer, agg.consumers[0])
}

func TestAggregator_Start_NoProvidersOrConsumers(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	err := agg.Start()
	assert.Error(t, err)
	assert.Equal(t, "no providers or consumers added", err.Error())
}

func TestAggregator_Start_ProviderConnectError(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	provider := &MockChatProvider{Name: "TestProvider", ConnectErr: errors.New("connect error")}
	agg.AddProvider(provider)
	agg.AddConsumer(&MockChatConsumer{Name: "TestConsumer"})
	err := agg.Start()
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond) // Allow goroutines to run
	assert.True(t, provider.Connected)
	agg.Stop()
}

func TestAggregator_Start_ProviderListenError(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	provider := &MockChatProvider{Name: "TestProvider", ListenErr: errors.New("listen error")}
	agg.AddProvider(provider)
	agg.AddConsumer(&MockChatConsumer{Name: "TestConsumer"})
	err := agg.Start()
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond) // Allow goroutines to run
	assert.True(t, provider.Connected)
	assert.True(t, provider.ListenCalled)
	agg.Stop()
}

func TestAggregator_Start_Success(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	provider := &MockChatProvider{Name: "TestProvider", Messages: []chatmodels.ChatMessage{{Provider: "TestProvider", Content: "Test Message"}}}
	consumer := &MockChatConsumer{Name: "TestConsumer"}
	agg.AddProvider(provider)
	agg.AddConsumer(consumer)
	err := agg.Start()
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond) // Allow goroutines to run
	assert.True(t, provider.Connected)
	assert.True(t, provider.ListenCalled)
	assert.Equal(t, 1, consumer.ConsumedCount)
	assert.True(t, consumer.StartCalled)
	assert.Equal(t, "Test Message", consumer.ConsumedMessages[0].Content)
	agg.Stop()
	assert.True(t, provider.Disconnected)
	assert.Nil(t, agg.stop)
}

func TestAggregator_Stop_NotStarted(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	agg.Stop()
	assert.Nil(t, agg.stop)
}

func TestAggregator_GetProvidersCount(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	agg.AddProvider(&MockChatProvider{Name: "TestProvider1"})
	agg.AddProvider(&MockChatProvider{Name: "TestProvider2"})
	assert.Equal(t, 2, agg.GetProvidersCount())
}

func TestAggregator_GetConsumersCount(t *testing.T) {
	agg := NewAggregator(&config.Config{})
	agg.AddConsumer(&MockChatConsumer{Name: "TestConsumer1"})
	agg.AddConsumer(&MockChatConsumer{Name: "TestConsumer2"})
	assert.Equal(t, 2, agg.GetConsumersCount())
}

func TestAggregator_MultipleProvidersAndConsumers(t *testing.T) {
	agg := NewAggregator(&config.Config{})

	provider1 := &MockChatProvider{Name: "Provider1", Messages: []chatmodels.ChatMessage{{Provider: "Provider1", Content: "Message from Provider1"}}}
	provider2 := &MockChatProvider{Name: "Provider2", Messages: []chatmodels.ChatMessage{{Provider: "Provider2", Content: "Message from Provider2"}}}
	consumer1 := &MockChatConsumer{Name: "Consumer1"}
	consumer2 := &MockChatConsumer{Name: "Consumer2"}

	agg.AddProvider(provider1)
	agg.AddProvider(provider2)
	agg.AddConsumer(consumer1)
	agg.AddConsumer(consumer2)

	err := agg.Start()
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assert.True(t, provider1.Connected)
	assert.True(t, provider2.Connected)
	assert.True(t, provider1.ListenCalled)
	assert.True(t, provider2.ListenCalled)

	assert.True(t, consumer1.StartCalled)
	assert.True(t, consumer2.StartCalled)
	assert.Equal(t, 2, consumer1.ConsumedCount)
	assert.Equal(t, 2, consumer2.ConsumedCount)

	assert.Contains(t, consumer1.ConsumedMessages, chatmodels.ChatMessage{Provider: "Provider1", Content: "Message from Provider1"})
	assert.Contains(t, consumer1.ConsumedMessages, chatmodels.ChatMessage{Provider: "Provider2", Content: "Message from Provider2"})
	assert.Contains(t, consumer2.ConsumedMessages, chatmodels.ChatMessage{Provider: "Provider1", Content: "Message from Provider1"})
	assert.Contains(t, consumer2.ConsumedMessages, chatmodels.ChatMessage{Provider: "Provider2", Content: "Message from Provider2"})

	agg.Stop()
	time.Sleep(200 * time.Millisecond)
	assert.True(t, provider1.Disconnected)
	assert.True(t, provider2.Disconnected)
}
