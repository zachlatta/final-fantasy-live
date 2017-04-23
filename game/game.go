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
	actionInterval = 15

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

	playerOne ui.ControllerAdapter
	playerTwo ui.ControllerAdapter

	buttonsToPress chan int

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

		buttonsToPress: make(chan int),

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}
}

func (g Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	go g.listenForReactions()
	go g.handleButtonPresses()

	// Emulator must be on main thread
	g.startEmulator()
}

func (g Game) listenForReactions() {
	ticker := time.NewTicker(1 * time.Second)
	timer := actionInterval

	for range ticker.C {
		fmt.Printf("%d...\n", timer)
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
		return nes.ButtonUp
	case facebook.ReactionLove:
		return nes.ButtonDown
	case facebook.ReactionHaha:
		return nes.ButtonLeft
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

func (g Game) handleButtonPresses() {
	for action := range g.buttonsToPress {
		fmt.Printf("Pressing %s....\n", buttonToString[action])
		g.playerOne.Trigger(action, true)
		time.Sleep(buttonPressTime)
		g.playerOne.Trigger(action, false)
	}
}

func (g Game) startEmulator() {
	emulator.Emulate(
		g.RomPath,
		g.playerOne,
		g.playerTwo,
	)
}
