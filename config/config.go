package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ConnectTwitch               bool
	TwitchChannel               string
	ConnectYoutube              bool
	YoutubeApiKey               string
	YoutubeChannelId            string
	YoutubeQueriesPerDay        int
	ChatOutput                  bool
	WebpageOutput               bool
	WebpageOutputPort           int
	WebpageOuputShortenProvider bool
	WebpageOutputHideProvider   bool
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
			log.Println("No number of Youtube queries per day found, using default limit:", defaultYoutubeQueriesPerDay)
			youtubeQueriesPerDay = defaultYoutubeQueriesPerDay
		}

		outputChat, _ := strconv.ParseBool(os.Getenv("OUTPUT_CHAT"))
		webpageOutput, _ := strconv.ParseBool(os.Getenv("OUTPUT_WEBPAGE"))
		webpageOutputPort, err := strconv.Atoi(os.Getenv("OUTPUT_WEBPAGE_PORT"))

		if (err != nil || webpageOutputPort == 0) && webpageOutput {
			webpageOutputPort = 8080
		}

		webpageOuputShortenProvider, err := strconv.ParseBool(os.Getenv("OUTPUT_WEBPAGE_SHORTEN_PROVIDER"))

		if err != nil && webpageOutput {
			webpageOuputShortenProvider = false
		}

		webpageOutputHideProvider, err := strconv.ParseBool(os.Getenv("OUTPUT_WEBPAGE_HIDE_PROVIDER"))

		if err != nil && webpageOutput {
			webpageOutputHideProvider = false
		}

		defaultWebpageOutputPort := 8080
		if err != nil && webpageOutput {
			log.Println("No port specified for webpage output, using default:", defaultWebpageOutputPort)
			webpageOutputPort = defaultWebpageOutputPort
		}

		config = &Config{
			ConnectTwitch:               connectTwitch,
			TwitchChannel:               os.Getenv("TWITCH_CHANNEL"),
			ConnectYoutube:              connectYoutube,
			YoutubeApiKey:               os.Getenv("YOUTUBE_API_KEY"),
			YoutubeChannelId:            os.Getenv("YOUTUBE_CHANNEL_ID"),
			YoutubeQueriesPerDay:        youtubeQueriesPerDay,
			ChatOutput:                  outputChat,
			WebpageOutput:               webpageOutput,
			WebpageOutputPort:           webpageOutputPort,
			WebpageOuputShortenProvider: webpageOuputShortenProvider,
			WebpageOutputHideProvider:   webpageOutputHideProvider,
		}
	})
	return config
}
