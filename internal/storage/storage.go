package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Storage struct {
	path string
}

type SaveVideoStatOptions struct {
	Id      string
	Title   string
	Channel VideoChannel
	Views   int
}

type VideoViews struct {
	Views     int
	Timestamp int64
}

type VideoChannel struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type VideoStatDetails struct {
	Id             string       `json:"id"`
	Title          string       `json:"title"`
	Channel        VideoChannel `json:"channel"`
	LastActivityAt int64        `json:"last_activity_at"`
	Views          []VideoViews `json:"views"`
}

func NewStorage() (*Storage, error) {
	rootPath, err := findProjectRoot()
	if err != nil {
		return &Storage{}, err
	}

	return &Storage{path: fmt.Sprintf("%s/%s/", rootPath, "storage")}, nil
}

func (s *Storage) SaveVideoStats(options SaveVideoStatOptions) (VideoStatDetails, error) {
	file := fmt.Sprintf("%s%s.%s", s.path, options.Id, "json")

	var config VideoStatDetails

	// Check if the file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// Create a new Viper configuration
		viper.SetConfigFile(file)

		// Set default
		viper.SetDefault("id", options.Id)
		viper.SetDefault("title", options.Title)
		viper.SetDefault("last_activity_at", time.Now().Unix())
		viper.SetDefault("channel", VideoChannel{Id: options.Channel.Id, Title: options.Channel.Title})
		viper.SetDefault("views", []VideoViews{{Views: options.Views, Timestamp: time.Now().Unix()}})

		// Save the configuration file with default values
		if err := viper.WriteConfigAs(file); err != nil {
			return VideoStatDetails{}, err
		}
	} else {
		// Configuration file exists, read it
		viper.SetConfigFile(file)

		if err := viper.ReadInConfig(); err != nil {
			return VideoStatDetails{}, err
		}

		// Unmarshal the configuration into the struct
		if err := viper.Unmarshal(&config); err != nil {
			return VideoStatDetails{}, nil
		}

		// Append the views
		views := append(config.Views, VideoViews{Views: options.Views, Timestamp: time.Now().Unix()})
		viper.Set("views", views)
		viper.Set("last_activity_at", time.Now().Unix())

		// Save the updated configuration to the file
		if err := viper.WriteConfig(); err != nil {
			return VideoStatDetails{}, err
		}
	}

	// Read the config again
	if err := viper.ReadInConfig(); err != nil {
		return VideoStatDetails{}, err
	}

	// Unmarshal the configuration into the struct
	if err := viper.Unmarshal(&config); err != nil {
		return VideoStatDetails{}, nil
	}

	return config, nil
}

func findProjectRoot() (string, error) {
	// Start from the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Define the name of the marker file or directory
	markerName := "main.go" // Change this to your marker's name

	// Traverse up the directory tree until we find the marker
	for {
		// Check if the marker file or directory exists
		markerPath := filepath.Join(cwd, markerName)
		_, err := os.Stat(markerPath)
		if err == nil {
			return cwd, nil // Found the marker, so this is the project's root
		}

		// Move up one directory level
		parent := filepath.Dir(cwd)
		if parent == cwd {
			// Reached the root directory without finding the marker
			return "", fmt.Errorf("marker %s not found", markerName)
		}
		cwd = parent
	}
}
