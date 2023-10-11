/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/ricomonster/black-flag/internal/videos"
	"github.com/spf13/cobra"
)

// viewCmd represents the view command
var (
	showViews bool
	viewCmd   = &cobra.Command{
		Use: "view",
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

			videoItem, err := videoLib.FindVideo(video)
			if err != nil {
				fmt.Printf("Something went wrong %v", err)
				os.Exit(0)
			}

			// Check if the video exists
			if videoItem.Id == "" {
				fmt.Printf("Video not yet included to our list. Run \"add --video=%s\"\n", video)
				os.Exit(0)
			}

			// Process the video
			videoItem, err = videoLib.ProcessVideoStat(video)
			if err != nil {
				fmt.Printf("Something went wrong %v", err)
				os.Exit(0)
			}

			renderVideoDetails(videoItem)
		},
	}
)

func renderVideoHistory(video videos.VideoDdbAttributes) {
}

// Will render/show the basic video details and some comparison of the recorded view stat.
func renderVideoDetails(video videos.VideoDdbAttributes) {
	// Get the last item in ViewLogs
	lastIndex := len(video.ViewLogs) - 1
	lastItem := video.ViewLogs[lastIndex]

	var beforeLastItem videos.VideoViewAttributes

	// Check if there's previous item after the last item
	beforeLastIndex := lastIndex - 1
	if beforeLastIndex > 0 && beforeLastIndex < len(video.ViewLogs) {
		beforeLastItem = video.ViewLogs[beforeLastIndex]
	}

	// Display something
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle(video.Title)

	if beforeLastItem.Views == 0 {
		t.AppendRow(table.Row{"Current:", lastItem.Views})
	} else {
		t.AppendRows([]table.Row{
			{"Current:", lastItem.Views, "Added:", lastItem.Views - beforeLastItem.Views},
			{"Previous", beforeLastItem.Views, "Last Run:", time.Unix(beforeLastItem.Timestamp, 0).Local()},
		})
	}

	t.Render()
}

func init() {
	rootCmd.AddCommand(viewCmd)

	// view --video=url
	viewCmd.Flags().StringP("video", "v", "", "Video ID or URL of the youtube video to view.")
	_ = viewCmd.MarkFlagRequired("video")

	// view --show-history
	viewCmd.Flags().BoolVarP(&showViews, "show-views", "s", true, "Shows the stored view logs for the video.")
}
