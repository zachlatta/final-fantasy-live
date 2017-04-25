package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zachlatta/nostalgic-rewind/facebook"
	"github.com/zachlatta/nostalgic-rewind/game"
	"github.com/zachlatta/nostalgic-rewind/util"
)

var accessToken string
var romPath string
var vidId string
var vidStreamUrl string
var savePath string

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Manage a Facebook Live stream",
}

var createStreamCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Facebook Live stream",
	Run: func(cmd *cobra.Command, args []string) {
		if accessToken == "" {
			fmt.Fprintln(os.Stderr, "Access token is required.")
			os.Exit(1)
		}

		vid, err := facebook.CreateLiveVideo(accessToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating stream:", err)
			os.Exit(1)
		}

		fmt.Println("Stream created!")
		fmt.Println()
		fmt.Println("ID:", vid.Id)
		fmt.Println("Stream URL:", vid.StreamUrl)
		fmt.Println()
		fmt.Println("Run `stream play` to cast to the stream.")
	},
}

var playStreamCmd = &cobra.Command{
	Use:   "play [path to rom to play]",
	Short: "Start casting the given ROM",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Please provide a path to the ROM to play.")
			os.Exit(1)
		}

		if accessToken == "" || vidId == "" || vidStreamUrl == "" {
			fmt.Fprintln(os.Stderr, "Access token, stream ID, and stream URL are all required.")
			os.Exit(1)
		}

		romPath := args[0]
		vid := facebook.LiveVideo{
			Id:        vidId,
			StreamUrl: vidStreamUrl,
		}

		fmt.Printf("Starting %s...\n", romPath)

		spath := filepath.Join(savePath, util.MD5HashString(romPath), game.GameSavePath)

		var g game.Game

		ok, err := util.FileExists(spath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Wasn't able to check if a file existed", err)
			os.Exit(1)
		}

		if ok {
			fmt.Println("Loading game from save...")

			var save game.Save
			f, err := os.Open(spath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error loading save", err)
				os.Exit(1)
			}

			if err := json.NewDecoder(f).Decode(&save); err != nil {
				fmt.Fprintln(os.Stderr, "Error decoding game save", err)
				os.Exit(1)
			}

			g, err = game.NewFromSave(save, vid, accessToken)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating game:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Loading game without existing save")
			g, err = game.New(vid, romPath, accessToken, savePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating game:", err)
				os.Exit(1)
			}
		}

		g.Start()
	},
}

func init() {
	RootCmd.AddCommand(streamCmd)

	streamCmd.AddCommand(createStreamCmd)
	createStreamCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Facebook access token for page to stream from")

	streamCmd.AddCommand(playStreamCmd)
	playStreamCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Facebook access token for page to stream from")
	playStreamCmd.Flags().StringVarP(&vidId, "stream-id", "i", "", "ID of Facebook Live stream to cast to")
	playStreamCmd.Flags().StringVarP(&vidStreamUrl, "stream-url", "u", "", "URL of Facebook Live stream to cast to")
	playStreamCmd.Flags().StringVarP(&savePath, "save", "s", "./.saves", "The directory to save the state of the emulator")
}
