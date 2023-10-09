/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/ricomonster/black-flag/cmd"
	"github.com/ricomonster/black-flag/internal/config"
)

func init() {
	// Load the env
	config.LoadEnvConfig()
}

func main() {
	cmd.Execute()
}
