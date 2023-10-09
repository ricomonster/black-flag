/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ricomonster/black-flag/internal/videos"
	"github.com/ricomonster/black-flag/internal/youtube"
	"github.com/spf13/cobra"
)

// viewCmd represents the view command
var (
	stats   bool
	viewCmd = &cobra.Command{
		Use: "view",
		Run: func(cmd *cobra.Command, _ []string) {
			video := cmd.Flags().Lookup("video").Value.String()

			if strings.Contains(video, "?") {
				// Split
				splitString := strings.Split(video, "v=")
				video = splitString[1]
			}

			// Instantiate the video lib
			videoLib := videos.NewVideos()
			ddbItem, err := videoLib.FindVideo(video)
			if err != nil {
				fmt.Printf("Something went wrong %v", err)
				os.Exit(0)
			}

			// Check if the video exists
			if ddbItem.Id == "" {
				fmt.Printf("Video not yet included to our list. Run \"add --video=%s\"\n", video)
				os.Exit(0)
			}

			// Stat fetching is enabled
			if stats {
				// Get video details from youtube
				svc, err := youtube.NewYoutube()
				if err != nil {
					fmt.Printf("Something went wrong %v", err)
					os.Exit(0)
				}

				videoDetails, err := svc.GetVideoDetails(video)
				if err != nil {
					fmt.Printf("Something went wrong %v", err)
					os.Exit(0)
				}

				// Setup the data
				videoData := videos.SaveVideoOptions{
					Id:    videoDetails.Id,
					Title: videoDetails.Snippet.Title,
					Views: int(videoDetails.Statistics.ViewCount),
					Channel: videos.VideoChannelAttributes{
						Id:    videoDetails.Snippet.ChannelId,
						Title: videoDetails.Snippet.ChannelTitle,
					},
				}

				// Save the video data
				err = videoLib.SaveVideo(videoData)
				if err != nil {
					fmt.Printf("Something went wrong %v", err)
					os.Exit(0)
				}

				// TODO: Fetch the video from DynamoDB again
			}

			// Display the details of the video
		},
	}
)

func init() {
	rootCmd.AddCommand(viewCmd)

	// view --video=url
	viewCmd.Flags().StringP("video", "v", "", "Video ID or URL of the youtube video to view.")
	_ = viewCmd.MarkFlagRequired("video")

	// view --stats
	viewCmd.Flags().BoolVarP(&stats, "stats", "s", true, "Will fetch the latest stats of the video.")
}
