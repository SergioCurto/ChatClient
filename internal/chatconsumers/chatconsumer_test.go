package chatconsumers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcreteChatConsumerFactory_CreateConsumer(t *testing.T) {
	factory := NewConcreteChatConsumerFactory()

	// Test creating a Console consumer
	consumer, err := factory.CreateConsumer(Console)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)
	assert.Equal(t, "Console", consumer.GetName())

	// Test creating a SimplePage consumer
	consumer, err = factory.CreateConsumer(SimplePage)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)
	assert.Equal(t, "SimplePage", consumer.GetName())
	
	// Test creating an unknown consumer
	consumer, err = factory.CreateConsumer(ChatConsumerType(999)) // Invalid consumer type
	assert.Error(t, err)
	assert.Nil(t, consumer)
}
