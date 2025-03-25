package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/aggregator"
	"github.com/SergioCurto/ChatClient/internal/chatproviders/twitch"
	"github.com/SergioCurto/ChatClient/internal/chatproviders/youtube"
)

func main() {
	fmt.Println("Chat client application started.")

	// Get Config
	cfg := config.GetConfig()

	agg := aggregator.NewAggregator(cfg)

	// Create and add providers configured
	if cfg.ConnectTwitch {
		fmt.Println("Creating and enabling Twitch chat provider")
		twitchProvider := twitch.NewTwitchProvider()
		agg.AddProvider(twitchProvider)
	}
	
	if cfg.ConnectYoutube {
		fmt.Println("Creating and enabling Youtube chat provider")
		youtubeProvider := youtube.NewYoutubeProvider()
		agg.AddProvider(youtubeProvider)
	}
	
	if agg.GetProvidersCount() == 0 {
		log.Fatal("No chat providers configured")
	}

	// Start the aggregator
	agg.Start()

	// Handle CTRL+C to gracefully shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		for msg := range agg.GetMessages() {
			fmt.Printf("[%s] %s\n", msg.Provider, msg.Content)
		}
		done <- true
	}()

	<-sigs
	fmt.Println("Shutting down...")
	agg.Stop()
	<-done

	fmt.Println("Chat aggregation ended.")
}
