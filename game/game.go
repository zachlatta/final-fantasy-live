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
	Emulator    *emulator.Emulator

	comments        chan facebook.Comment
	lastCommentTime time.Time
}

func New(vid facebook.LiveVideo, romPath string, accessToken string) (Game, error) {
	playerOne := ui.NewKeyboardControllerAdapter()
	playerTwo := &ui.DummyControllerAdapter{}

	e, err := emulator.NewEmulator(
		emulator.DefaultSettings,
		playerOne,
		playerTwo,
	)

	if err != nil {
		return Game{}, err
	}

	playerOne.SetWindow(e.Director.Window())

	return Game{
		Video:           vid,
		RomPath:         romPath,
		AccessToken:     accessToken,
		Emulator:        e,
		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}, nil
}

func (g Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	go g.listenForComments()
	go g.handleComments()

	// Emulator must be on main thread
	g.startEmulator()
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

func (g Game) handleComments() {
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
		case "start":
			action = nes.ButtonStart
		case "select":
			action = nes.ButtonSelect
		case "a":
			action = nes.ButtonA
		case "b":
			action = nes.ButtonB
		}

		if action != -1 {
			g.Emulator.PlayerOneController.Trigger(action, true)
			time.Sleep(time.Second / 5)
			g.Emulator.PlayerOneController.Trigger(action, false)
		}
	}
}

func (g Game) startEmulator() {
	g.Emulator.Play(g.RomPath)
}
