package aggregator

import (
	"errors"
	"fmt"
	"sync"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatconsumers"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/SergioCurto/ChatClient/internal/chatproviders"
)

type Aggregator struct {
	providers []chatproviders.ChatProvider
	consumers []chatconsumers.ChatConsumer
	messages  chan chatmodels.ChatMessage
	cfg       *config.Config
	stop      chan struct{}
	wg        sync.WaitGroup
}

func NewAggregator(cfg *config.Config) *Aggregator {
	return &Aggregator{
		messages: make(chan chatmodels.ChatMessage),
		cfg:      cfg,
	}
}

func (a *Aggregator) AddProvider(provider chatproviders.ChatProvider) {
	a.providers = append(a.providers, provider)
}

func (a *Aggregator) AddConsumer(consumer chatconsumers.ChatConsumer) {
	a.consumers = append(a.consumers, consumer)
}

func (a *Aggregator) Start() error {
	if a.stop != nil {
		return errors.New("aggregator already started")
	}

	if (len(a.providers) == 0) || (len(a.consumers) == 0) {
		return errors.New("no providers or consumers added")
	}

	a.messages = make(chan chatmodels.ChatMessage)

	a.stop = make(chan struct{})

	for _, provider := range a.providers {
		a.wg.Add(1)
		go func(p chatproviders.ChatProvider) {
			defer a.wg.Done()

			err := p.Connect(a.cfg)
			if err != nil {
				fmt.Println("Error connecting to provider:", p.GetName(), err)
				return
			}
			err = p.Listen(a.messages)
			if err != nil {
				fmt.Println("Error on provider:", p.GetName(), err)
			}
			<-a.stop // Wait for the stop signal before disconnecting
			fmt.Println("Stopping provider:", p.GetName())
			p.Disconnect()
		}(provider)
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		fmt.Println("Starting consumers...")
		for _, consumer := range a.consumers {
			go func() {
				err := consumer.Start(a.cfg)
				if err != nil {
					fmt.Println("Ignoring consumer with error during start:", consumer.GetName(), err)
				}
			}()
		}

		fmt.Println("Consumers started")

		for msg := range a.messages {
			for _, consumer := range a.consumers {
				go consumer.Consume(msg)
			}
		}
	}()

	return nil
}

func (a *Aggregator) Stop() {
	if a.stop != nil {
		close(a.stop)
		a.stop = nil
		close(a.messages) // Close the messages channel to evict the consumers
		a.wg.Wait()
	}
}

func (a *Aggregator) GetProvidersCount() int {
	return len(a.providers)
}

func (a *Aggregator) GetConsumersCount() int {
	return len(a.consumers)
}
