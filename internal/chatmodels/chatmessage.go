package chatmodels

import "time"

type ChatMessage struct {
	Provider          string
	ProviderShortName string
	Timestamp         time.Time
	Content           string
	AuthorName        string
}
