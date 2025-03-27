package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/aggregator"
	"github.com/SergioCurto/ChatClient/internal/chatconsumers"
	"github.com/SergioCurto/ChatClient/internal/chatproviders"
)

func main() {
	fmt.Println("Chat client application started.")

	// Get Config
	cfg := config.GetConfig()

	agg := aggregator.NewAggregator(cfg)
	chatProviderFactory := chatproviders.NewConcreteChatProviderFactory()

	// Create and add providers configured
	if cfg.ConnectTwitch {
		fmt.Println("Creating and enabling Twitch chat provider")
		twitchProvider, err := chatProviderFactory.CreateProvider(chatproviders.Twitch)
		if err != nil {
			log.Fatal("Error creating Twitch provider: ", err)
		}
		agg.AddProvider(twitchProvider)
	}

	if cfg.ConnectYoutube {
		fmt.Println("Creating and enabling Youtube chat provider")
		youtubeProvider, err := chatProviderFactory.CreateProvider(chatproviders.Youtube)
		if err != nil {
			log.Fatal("Error creating Youtube provider: ", err)
		}
		agg.AddProvider(youtubeProvider)
	}

	// Create and add consumers configured
	if cfg.ChatOutput {
		fmt.Println("Creating and enabling Console consumer")
		consumerFactory := chatconsumers.NewConcreteChatConsumerFactory()
		consumer, err := consumerFactory.CreateConsumer(chatconsumers.Console)
		if err != nil {
			log.Fatal("Error creating Console consumer: ", err)
		}
		agg.AddConsumer(consumer)
	}
	
	// Start the aggregator
	err := agg.Start()

	if err != nil {
		log.Fatal("Error starting aggregator: ", err)
	}

	// Handle CTRL+C to gracefully shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	fmt.Println("Shutting down...")
	agg.Stop()

	fmt.Println("Chat aggregation ended.")
}
