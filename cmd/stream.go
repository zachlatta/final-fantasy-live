package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zachlatta/nostalgic-rewind/facebook"
	"github.com/zachlatta/nostalgic-rewind/game"
)

var accessToken string
var romPath string

var streamCmd = &cobra.Command{
	Use:   "stream [path to rom to play]",
	Short: "Start a Facebook Live stream.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Please provide the name of the game to stream.")
			os.Exit(1)
		}

		if accessToken == "" {
			fmt.Fprintln(os.Stderr, "Access token is required.")
			os.Exit(1)
		}

		romPath := args[0]

		fmt.Printf("Starting %s...\n", romPath)

		vid, err := facebook.CreateLiveVideo(accessToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating stream:", err)
			os.Exit(1)
		}

		game := game.New(vid, romPath, accessToken)

		game.Start()
	},
}

func init() {
	RootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Facebook access token for page to stream from")
}
