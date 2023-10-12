package videos

import (
	"strings"
	"time"

	"github.com/ricomonster/black-flag/internal/aws/dynamodb"
	"github.com/ricomonster/black-flag/internal/youtube"
)

type VideoDdbAttributes struct {
	Id                string                 `dynamodbav:"Id"`
	Title             string                 `dynamodbav:"Title"`
	Channel           VideoChannelAttributes `dynamodbav:"Channel"`
	LastActivityAt    int64                  `dynamodbav:"LastActivityAt"`
	ViewLogs          []VideoViewAttributes  `dynamodbav:"ViewLogs"`
	Created           int64                  `dynamodbav:"Created"`
	Modified          int64                  `dynamodbav:"Modified"`
	TestAttribute     string                 `dynamodbav:"TestAttribute"`
	TestAttributeHana int                    `dynamodbav:"TestAttributeHana"`
}

type VideoViewAttributes struct {
	Views     int   `dynamodbav:"Views"`
	Timestamp int64 `dynamodbav:"Timestamp"`
}

type VideoChannelAttributes struct {
	Id    string
	Title string
}

type SaveVideoOptions struct {
	Id      string
	Title   string
	Channel VideoChannelAttributes
	Views   int
}

type videos struct {
	dynamodb *dynamodb.DynamoDB[VideoDdbAttributes]
	youtube  *youtube.Youtube
}

var TABLE = "BlackFlag_Videos"

func NewVideos() (*videos, error) {
	// Instantiate DynamoDB service
	ddbSvc := dynamodb.NewDynamoDB[VideoDdbAttributes](TABLE)

	// Instantiate youtube lib
	youtubeLib, err := youtube.NewYoutube()
	if err != nil {
		return &videos{}, err
	}

	return &videos{dynamodb: ddbSvc, youtube: youtubeLib}, nil
}

// Fetches all the stored videos records saved.
func (v *videos) GetVideos() ([]VideoDdbAttributes, error) {
	videos, err := v.dynamodb.GetAll()
	if err != nil {
		return []VideoDdbAttributes{}, err
	}

	return videos, nil
}

// Find a video record in the DynamoDB table
func (v *videos) FindVideo(id string) (VideoDdbAttributes, error) {
	// Check if the video already exists in DynamoDB
	item, err := v.dynamodb.FindById("Id", id)
	if err != nil {
		return VideoDdbAttributes{}, err
	}

	return item, nil
}

// Handles saving and updating video data to DynamoDB
func (v *videos) SaveVideo(options SaveVideoOptions) error {
	// Check if the video already exists so we'll just update it
	item, err := v.FindVideo(options.Id)
	if err != nil {
		return err
	}

	viewLog := VideoViewAttributes{Views: options.Views, Timestamp: time.Now().Unix()}

	if item.Id == "" {
		// Insert
		err := v.dynamodb.PutItem(VideoDdbAttributes{
			Id:             options.Id,
			Title:          options.Title,
			Channel:        options.Channel,
			LastActivityAt: time.Now().Unix(),
			ViewLogs:       []VideoViewAttributes{viewLog},
			Created:        time.Now().Unix(),
			Modified:       time.Now().Unix(),
		})
		if err != nil {
			return err
		}
	}

	// Update
	// Append to the ViewLog
	updatedViewLog := append(item.ViewLogs, viewLog)
	err = v.dynamodb.UpdateItem("Id", options.Id, VideoDdbAttributes{
		ViewLogs:       updatedViewLog,
		LastActivityAt: time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	return nil
}

// This will handle fetching of video stats from Youtube API and saving it to DynamoDB
func (v *videos) ProcessVideoStat(video string) (VideoDdbAttributes, error) {
	id := video
	if strings.Contains(id, "?") {
		// Split
		splitString := strings.Split(video, "v=")
		id = splitString[1]
	}

	// Get the video from dynamodb
	videoItem, err := v.FindVideo(id)
	if err != nil {
		return VideoDdbAttributes{}, err
	}

	// Check if video exists
	if videoItem.Id != "" {
		// Video already exists but before we proceed, let's make sure that we do not abuse the endpoint
		// so we need to check if the last activity for this record was more than 60 minutes
		currentTimestamp := time.Now().Unix()
		diff := currentTimestamp - videoItem.LastActivityAt
		if diff <= 3600 {
			return videoItem, nil
		}
	}

	// Get data from Youtube
	youtubeVideoData, err := v.youtube.GetVideoDetails(id)
	if err != nil {
		return VideoDdbAttributes{}, err
	}

	// Save data to dynamodb
	err = v.SaveVideo(SaveVideoOptions{
		Id:    youtubeVideoData.Id,
		Title: youtubeVideoData.Snippet.Title,
		Views: int(youtubeVideoData.Statistics.ViewCount),
		Channel: VideoChannelAttributes{
			Id:    youtubeVideoData.Snippet.ChannelId,
			Title: youtubeVideoData.Snippet.ChannelTitle,
		},
	})
	if err != nil {
		return VideoDdbAttributes{}, err
	}

	// Get again the data from dynamodb
	videoItem, err = v.FindVideo(id)
	if err != nil {
		return VideoDdbAttributes{}, err
	}

	return videoItem, nil
}
