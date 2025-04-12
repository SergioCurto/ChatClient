package youtube

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeProvider struct {
	Name          string
	ShortName     string
	apiKey        string
	channelId     string
	queriesPerDay int
	service       *youtube.Service
	liveChatId    string
	nextPage      string
	nextPoll      time.Duration
}

func NewYoutubeProvider() *YoutubeProvider {

	return &YoutubeProvider{
		Name:      "Youtube",
		ShortName: "Yt",
	}
}

func (y *YoutubeProvider) Connect(cfx *config.Config) error {
	fmt.Println("Connecting to Youtube with API Key...")

	y.apiKey = cfx.YoutubeApiKey
	y.channelId = cfx.YoutubeChannelId
	y.queriesPerDay = cfx.YoutubeQueriesPerDay

	if y.apiKey == "" || y.channelId == "" {
		return fmt.Errorf("missing YOUTUBE_API_KEY or YOUTUBE_CHANNEL_ID in environment variables")
	}

	// Create a new YouTube service client.
	ctx := context.Background()
	service, err := youtube.NewService(
		ctx,
		option.WithAPIKey(y.apiKey),
	)
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}
	y.service = service

	// Step 1: Find the Active Live Broadcast
	searchCall := y.service.Search.List([]string{"id", "snippet"}).
		ChannelId(y.channelId).
		EventType("live").
		Type("video").
		MaxResults(1)

	searchResponse, err := searchCall.Do()
	if err != nil {
		return fmt.Errorf("error searching for live broadcasts: %v", err)
	}

	if len(searchResponse.Items) == 0 {
		return fmt.Errorf("no active live broadcasts found for channel %s", y.channelId)
	}

	liveVideoId := searchResponse.Items[0].Id.VideoId
	if liveVideoId == "" {
		return fmt.Errorf("no live video id found for channel %s", y.channelId)
	}

	// Step 2: Get the Live Chat ID using the Live Video ID
	videoCall := y.service.Videos.List([]string{"liveStreamingDetails"}).
		Id(liveVideoId)

	videoResponse, err := videoCall.Do()
	if err != nil {
		return fmt.Errorf("error getting live video details: %v", err)
	}

	if len(videoResponse.Items) == 0 {
		return fmt.Errorf("no live video details found for video %s", liveVideoId)
	}

	if videoResponse.Items[0].LiveStreamingDetails == nil {
		return fmt.Errorf("no live streaming details found for video %s", liveVideoId)
	}

	y.liveChatId = videoResponse.Items[0].LiveStreamingDetails.ActiveLiveChatId
	if y.liveChatId == "" {
		return fmt.Errorf("no live chat found for video %s", liveVideoId)
	}

	fmt.Println("Live Chat ID:", y.liveChatId)

	return nil
}

func (y *YoutubeProvider) Disconnect() error {
	fmt.Println("Disconnecting from Youtube...")
	// No specific disconnect logic needed for YouTube API
	return nil
}

func (y *YoutubeProvider) Listen(messages chan<- chatmodels.ChatMessage) error {
	// Start listening for messages in a goroutine.
	go func(y *YoutubeProvider) {
		for {
			// Get the live chat messages.
			call := y.service.LiveChatMessages.List(y.liveChatId, []string{"snippet", "authorDetails"}).MaxResults(2000)
			if y.nextPage != "" {
				call = call.PageToken(y.nextPage)
			}
			response, err := call.Do()
			if err != nil {
				log.Printf("Error getting live chat messages: %v", err)
				return
			}

			y.nextPage = response.NextPageToken
			pollingInterval := time.Duration(response.PollingIntervalMillis) * time.Millisecond

			/* Youtube API is bad, and for multiple years did not implement a push based messaging system.
			   To overcome this we need to reduce the pooling rate based on the limits that the API key has.
			   See https://issuetracker.google.com/issues/35205195 */
			// Calculate the minimum polling interval based on queriesPerDay
			minPollingInterval := 24 * time.Hour / time.Duration(y.queriesPerDay)
			if pollingInterval < minPollingInterval {
				pollingInterval = minPollingInterval
			}
			y.nextPoll = pollingInterval

			// Process the messages.
			for _, item := range response.Items {

				messages <- chatmodels.ChatMessage{
					Provider:          y.GetName(),
					ProviderShortName: y.GetShortName(),
					Content:           item.Snippet.DisplayMessage,
					AuthorName:        item.AuthorDetails.DisplayName,
				}
			}
			// Wait for the next poll interval.
			time.Sleep(y.nextPoll)
		}
	}(y)
	return nil
}

func (y *YoutubeProvider) GetName() string {
	return y.Name
}

func (y *YoutubeProvider) GetShortName() string {
	return y.ShortName
}
