package youtube

import (
	"context"

	"google.golang.org/api/option"
	youtube_api "google.golang.org/api/youtube/v3"
)

type Youtube struct {
	service *youtube_api.Service
}

// Initialize the youtube internal library
func NewYoutube() (*Youtube, error) {
	// Initialize the YouTube Data API client
	ctx := context.Background()
	apiKey := "AIzaSyBrfauECMwGypNh8EMjPpq8u1T05i1VdlI"

	service, err := youtube_api.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return &Youtube{}, err
	}
	return &Youtube{service: service}, nil
}

// Returns the video details based on the video id.
func (yt *Youtube) GetVideoDetails(id string) (youtube_api.Video, error) {
	// Call the API to retrieve video details
	videoResponse, err := yt.service.Videos.List([]string{"snippet", "contentDetails", "statistics"}).Id(id).Do()
	if err != nil {
		return youtube_api.Video{}, err
	}

	return *videoResponse.Items[0], nil
}
