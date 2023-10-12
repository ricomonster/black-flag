/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/ricomonster/black-flag/internal/videos"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates and fetches video stats from Youtube API",
	Run: func(_ *cobra.Command, _ []string) {
		// Get the items from dynamodb
		// Instantiate the video lib
		videoLib, err := videos.NewVideos()
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			os.Exit(0)
		}

		allVideos, err := videoLib.GetVideos()
		if err != nil {
			fmt.Printf("Something went wrong %v", err)
			os.Exit(0)
		}

		// Create a channel to receive the results of the concurrent processing.
		results := make(chan videos.VideoDdbAttributes)

		// Setup goroutine
		var wg sync.WaitGroup
		wg.Add(len(allVideos))

		// Loop the videos
		for _, item := range allVideos {
			go func(item videos.VideoDdbAttributes) {
				defer wg.Done()
				// fmt.Printf("Processing %s\n", item.Title)

				// Process
				result, err := videoLib.ProcessVideoStat(item.Id)
				if err != nil {
					fmt.Printf("Error: %v", err)
					return
				}

				// Send the result to the channel.
				results <- result
			}(item)
		}

		// Close the channel when all of the goroutines have finished processing their items.
		go func() {
			wg.Wait()
			close(results)
		}()

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Title and Channel", "Views", "Added", "Last Run"})

		// Iterate over the channel to receive the results of the concurrent processing.
		for result := range results {
			// Get the last item in ViewLogs
			lastIndex := len(result.ViewLogs) - 1
			lastItem := result.ViewLogs[lastIndex]

			var beforeLastItem videos.VideoViewAttributes

			// Check if there's previous item after the last item
			beforeLastIndex := lastIndex - 1
			if beforeLastIndex > -1 && beforeLastIndex < len(result.ViewLogs) {
				beforeLastItem = result.ViewLogs[beforeLastIndex]
			}

			// First time
			added := 0
			lastRun := time.Unix(lastItem.Timestamp, 0).Local()

			if beforeLastItem.Views != 0 {
				added = lastItem.Views - beforeLastItem.Views
				lastRun = time.Unix(beforeLastItem.Timestamp, 0).Local()
			}

			t.AppendRow(table.Row{
				fmt.Sprintf("%s\n%s", result.Title, result.Channel.Title),
				lastItem.Views,
				added,
				lastRun,
			})
			t.AppendSeparator()
		}

		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
