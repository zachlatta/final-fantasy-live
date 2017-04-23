package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zachlatta/nostalgic-rewind/facebook"
)

var appId, appSecret string

var authenticateCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Get a long-lived access token",
	Run: func(cmd *cobra.Command, args []string) {
		if appId == "" || appSecret == "" {
			fmt.Fprintln(os.Stderr, "App ID and app secret required. See help. ")
			os.Exit(1)
		}

		token, err := facebook.Login(appId, appSecret)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error authenticating:", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("Access token:", token)
	},
}

func init() {
	RootCmd.AddCommand(authenticateCmd)
	authenticateCmd.Flags().StringVarP(&appId, "app-id", "i", "", "Facebook app ID")
	authenticateCmd.Flags().StringVarP(&appSecret, "app-secret", "s", "", "Facebook app secret")
}
