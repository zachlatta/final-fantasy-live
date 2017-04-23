package game

import (
	"fmt"
	"os"
	"time"

	"github.com/paked/nes/nes"
	"github.com/paked/nes/ui"
	"github.com/zachlatta/nostalgic-rewind/emulator"
	"github.com/zachlatta/nostalgic-rewind/facebook"
)

const (
	reactPollInterval = 2 * time.Second
)

type Game struct {
	Video       facebook.LiveVideo
	RomPath     string
	AccessToken string

	playerOne ui.ControllerAdapter
	playerTwo ui.ControllerAdapter

	comments        chan facebook.Comment
	lastCommentTime time.Time
}

func New(vid facebook.LiveVideo, romPath string, accessToken string) Game {
	return Game{
		Video:       vid,
		RomPath:     romPath,
		AccessToken: accessToken,

		playerOne: &ui.BasicControllerAdapter{},
		playerTwo: &ui.DummyControllerAdapter{},

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}
}

func (g Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	go g.listenForComments()
	go g.startEmulator()

	for comment := range g.comments {
		fmt.Println("New comment:", comment.Message)

		action := -1

		switch comment.Message {
		case "left":
			action = nes.ButtonLeft
		case "right":
			action = nes.ButtonRight
		case "up":
			action = nes.ButtonUp
		case "down":
			action = nes.ButtonDown
		}

		if action != -1 {
			g.playerOne.Trigger(action, true)
			time.Sleep(1 * time.Second)
			g.playerOne.Trigger(action, false)
		}
	}
}

func (g Game) listenForComments() {
	ticker := time.NewTicker(reactPollInterval)

	for range ticker.C {
		comments, err := facebook.Comments(g.Video.Id, g.AccessToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error listening for comments:", err)
			os.Exit(1)
		}

		for _, comment := range comments {
			if comment.Created.After(g.lastCommentTime) {
				g.comments <- comment
				g.lastCommentTime = comment.Created
			}
		}
	}
}

func (g Game) startEmulator() {
	emulator.Emulate(
		g.RomPath,
		g.playerOne,
		g.playerTwo,
	)
}
