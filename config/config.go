package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ConnectTwitch        bool
	TwitchChannel        string
	ConnectYoutube       bool
	YoutubeApiKey        string
	YoutubeChannelId     string
	YoutubeQueriesPerDay int
}

var config *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		//load .env file if it exists
		if err := godotenv.Load(); err != nil {
			log.Fatal("No .env file found")
		}

		connectTwitch, _ := strconv.ParseBool(os.Getenv("CONNECT_TWITCH"))
		
		defaultYoutubeQueriesPerDay := 10000
		connectYoutube, _ := strconv.ParseBool(os.Getenv("CONNECT_YOUTUBE"))
		youtubeQueriesPerDay, err := strconv.Atoi(os.Getenv("YOUTUBE_QUERIES_PER_DAY"))

		if err != nil && connectYoutube {
			log.Println("No number of Youtube queries per day found, using default limit", defaultYoutubeQueriesPerDay)
			youtubeQueriesPerDay = defaultYoutubeQueriesPerDay
		}

		config = &Config{
			ConnectTwitch:        connectTwitch,
			TwitchChannel:        os.Getenv("TWITCH_CHANNEL"),
			ConnectYoutube:       connectYoutube,
			YoutubeApiKey:        os.Getenv("YOUTUBE_API_KEY"),
			YoutubeChannelId:     os.Getenv("YOUTUBE_CHANNEL_ID"),
			YoutubeQueriesPerDay: youtubeQueriesPerDay,
		}
	})
	return config
}
