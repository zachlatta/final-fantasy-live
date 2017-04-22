package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var accessToken string

var streamCmd = &cobra.Command{
	Use:   "stream [game]",
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

		gameName := args[0]

		fmt.Printf("Starting %s...\n", gameName)
	},
}

func init() {
	RootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Facebook access token for page to stream from")
}
