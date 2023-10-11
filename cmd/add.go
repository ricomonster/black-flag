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

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a youtube video to get the stats.",
	Run: func(cmd *cobra.Command, _ []string) {
		video := cmd.Flags().Lookup("video").Value.String()

		if strings.Contains(video, "?") {
			// Split
			splitString := strings.Split(video, "v=")
			video = splitString[1]
		}

		// Instantiate the video lib
		videoLib, err := videos.NewVideos()
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			os.Exit(0)
		}

		ddbItem, err := videoLib.FindVideo(video)
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			os.Exit(0)
		}

		// Will only perform addition if the video is not yet included in our list
		if ddbItem.Id == "" {
			fmt.Printf("Fetching video details %s...\n", video)

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
		}

		// Fetch again
		ddbItem, err = videoLib.FindVideo(video)
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			os.Exit(0)
		}

		fmt.Printf("Video \"%s\" was already added.\nPlease run \"view --video=url/id\" to check its details.\n", ddbItem.Title)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// add --video
	addCmd.Flags().StringP("video", "v", "", "Video ID or URL of the youtube video to include.")

	_ = addCmd.MarkFlagRequired("video")
}
