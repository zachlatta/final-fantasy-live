package cmd

import (
	"fmt"
	"os"

	"github.com/paked/nes/nes"
	"github.com/paked/nes/ui"
	"github.com/spf13/cobra"
	"github.com/zachlatta/nostalgic-rewind/emulator"
)

var accessToken string
var romPath string

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

		exampleStartMario()
	},
}

func init() {
	RootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Facebook access token for page to stream from")
	streamCmd.Flags().StringVarP(&romPath, "rom", "r", "", "Path for the ROM to play")
}

func exampleStartMario() {
	// Player One is always holding down left joystick
	playerOne := &ui.BasicControllerAdapter{}
	playerOne.Trigger(nes.ButtonLeft, true)

	emulator.Emulate(
		romPath,
		playerOne,
		&ui.DummyControllerAdapter{},
	)
}
