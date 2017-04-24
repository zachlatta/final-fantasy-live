package game

import (
	"fmt"
	"os"
	"time"

	"github.com/paked/nes/nes"
	"github.com/paked/nes/ui"
	"github.com/zachlatta/nostalgic-rewind/emulator"
	"github.com/zachlatta/nostalgic-rewind/facebook"
	"github.com/zachlatta/nostalgic-rewind/obs"
	"github.com/zachlatta/nostalgic-rewind/util"
)

const (
	actionInterval = 10

	buttonPressTime = 1 * time.Second
)

var buttonToString = map[int]string{
	nes.ButtonUp:     "up",
	nes.ButtonDown:   "down",
	nes.ButtonRight:  "right",
	nes.ButtonLeft:   "left",
	nes.ButtonA:      "A",
	nes.ButtonB:      "B",
	nes.ButtonStart:  "start",
	nes.ButtonSelect: "select",
}

type Game struct {
	Video       facebook.LiveVideo
	RomPath     string
	AccessToken string
	Emulator    *emulator.Emulator
	Obs         obs.Obs

	buttonsToPress chan int

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

	streamUrl, streamKey := util.SplitStreamUrl(vid.StreamUrl)

	return Game{
		Video:       vid,
		RomPath:     romPath,
		AccessToken: accessToken,
		Emulator:    e,
		Obs:         obs.New(streamUrl, streamKey),

		buttonsToPress: make(chan int),

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}, nil
}

func (g Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	go g.startObs()
	go g.listenForReactions()
	go g.handleButtonPresses()

	// Emulator must be on main thread
	g.startEmulator()
}

func (g *Game) startObs() {
	if err := g.Obs.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running OBS:", err)
		os.Exit(1)
	}
}

func (g *Game) listenForReactions() {
	ticker := time.NewTicker(1 * time.Second)
	timer := actionInterval

	for range ticker.C {
		g.Obs.UpdateNextButtonPress(timer)
		timer -= 1

		if timer == 0 {
			timer = actionInterval

			reactions, err := facebook.Reactions(g.Video.Id, g.AccessToken)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error listening for reactions:", err)
				os.Exit(1)
			}

			if len(reactions) == 0 {
				fmt.Println("No reactions. Skipping button press.")
				continue
			}

			mostCommonReact := mostCommonReact(reactions)

			g.buttonsToPress <- reactionToButton(mostCommonReact)
		}
	}
}

func reactionToButton(reaction facebook.ReactionType) int {
	switch reaction {
	case facebook.ReactionLike:
		return nes.ButtonLeft
	case facebook.ReactionLove:
		return nes.ButtonUp
	case facebook.ReactionHaha:
		return nes.ButtonDown
	case facebook.ReactionWow:
		return nes.ButtonRight
	case facebook.ReactionSad:
		return nes.ButtonB
	case facebook.ReactionAngry:
		return nes.ButtonA
	default:
		return -1
	}
}

// Returns the most common reaction given an array of them
func mostCommonReact(reactions []facebook.Reaction) facebook.ReactionType {
	if len(reactions) == 0 {
		return -1
	}

	reactionCounts := map[facebook.ReactionType]int{}

	for _, reaction := range reactions {
		reactionCounts[reaction.Type] += 1
	}

	var mostCommon facebook.ReactionType
	var maxCount int

	for reactionType, occurances := range reactionCounts {
		if occurances > maxCount {
			mostCommon = reactionType
			maxCount = occurances
		}
	}

	return mostCommon
}

func (g *Game) handleButtonPresses() {
	for action := range g.buttonsToPress {
		g.Obs.IncrementButtonPresses()
		g.Obs.AddMostRecentPress(buttonToString[action])

		g.Emulator.PlayerOneController.Trigger(action, true)
		time.Sleep(time.Second / 5)
		g.Emulator.PlayerOneController.Trigger(action, false)
	}
}

func (g Game) startEmulator() {
	g.Emulator.Play(g.RomPath)
}
