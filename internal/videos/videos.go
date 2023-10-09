package videos

import (
	"time"

	"github.com/ricomonster/black-flag/internal/aws/dynamodb"
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
}

var TABLE = "BlackFlag_Videos"

func NewVideos() *videos {
	// Instantiate DynamoDB service
	ddbSvc := dynamodb.NewDynamoDB[VideoDdbAttributes](TABLE)
	return &videos{dynamodb: ddbSvc}
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
