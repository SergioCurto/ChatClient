package chatproviders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcreteChatProviderFactory_CreateProvider(t *testing.T) {
	factory := NewConcreteChatProviderFactory()

	// Test creating a Twitch provider
	provider, err := factory.CreateProvider(Twitch)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Twitch", provider.GetName())

	// Test creating a Youtube provider
	provider, err = factory.CreateProvider(Youtube)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Youtube", provider.GetName())

	// Test creating an unknown provider
	provider, err = factory.CreateProvider(ChatProviderType(999)) // Invalid provider type
	assert.Error(t, err)
	assert.Nil(t, provider)
}
