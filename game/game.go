package game

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	pollInterval   = 2 * time.Second

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
	SavePath    string
	AccessToken string
	Emulator    *emulator.Emulator
	Obs         obs.Obs

	startTime      time.Time
	reactionCounts map[facebook.ReactionType]int

	buttonsToPress chan int

	comments        chan facebook.Comment
	lastCommentTime time.Time
}

func New(vid facebook.LiveVideo, romPath string, accessToken string, savePath string) (Game, error) {
	playerOne := ui.NewKeyboardControllerAdapter()
	playerTwo := &ui.DummyControllerAdapter{}

	e, err := emulator.NewEmulator(
		emulator.DefaultSettings,
		playerOne,
		playerTwo,
		nesSaveFilePath(savePath, romPath),
	)

	if err != nil {
		return Game{}, err
	}

	streamUrl, streamKey := util.SplitStreamUrl(vid.StreamUrl)

	return Game{
		Video:       vid,
		RomPath:     romPath,
		SavePath:    savePath,
		AccessToken: accessToken,
		Emulator:    e,
		Obs:         obs.New(streamUrl, streamKey),

		reactionCounts: map[facebook.ReactionType]int{},

		buttonsToPress: make(chan int),

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}, nil
}

func (g *Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	g.startTime = time.Now()

	go g.startObs()
	go g.pollForReactions()
	go g.buttonCountdown()
	go g.handleButtonPresses()
	go g.continuoslySave()

	// Emulator must be on main thread
	g.startEmulator()
}

func (g *Game) Save() error {
	err := g.Emulator.SaveState(g.SavePath)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) Load() error {
	err := g.Emulator.LoadState(g.SavePath)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) startObs() {
	if err := g.Obs.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running OBS:", err)
		os.Exit(1)
	}
}

func (g *Game) pollForReactions() {
	ticker := time.NewTicker(pollInterval)

	for range ticker.C {
		reactions, err := facebook.Reactions(g.Video.Id, g.AccessToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error polling for reactions:", err)
			os.Exit(1)
		}

		g.reactionCounts = reactionCounts(reactions)

		// Update vote breakdown
		buttonVoteMap := map[int]int{}
		for reactionType, count := range g.reactionCounts {
			buttonVoteMap[reactionToButton(reactionType)] = count
		}
		if err := g.Obs.UpdateVoteBreakdown(buttonVoteMap); err != nil {
			fmt.Fprintln(os.Stderr, "Error updating vote breakdown:", err)
			os.Exit(1)
		}
	}
}

func (g *Game) buttonCountdown() {
	ticker := time.NewTicker(1 * time.Second)
	timer := actionInterval

	for range ticker.C {
		g.Obs.UpdateNextButtonPress(timer)
		g.Obs.UpdateTotalUptime(g.startTime, time.Now())

		timer -= 1

		if timer == 0 {
			timer = actionInterval

			mostCommonReact := mostCommonReact(g.reactionCounts)
			if mostCommonReact == -1 {
				fmt.Println("No reactions. Skipping button press.")
				continue
			}

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

func reactionCounts(reactions []facebook.Reaction) map[facebook.ReactionType]int {
	countMap := map[facebook.ReactionType]int{}
	for _, reaction := range reactions {
		countMap[reaction.Type] += 1
	}

	return countMap
}

// Returns the most common reaction given an array of them
func mostCommonReact(countMap map[facebook.ReactionType]int) facebook.ReactionType {
	var foundMostCommon bool
	var mostCommon facebook.ReactionType
	var maxCount int

	for reactionType, occurances := range countMap {
		if occurances > maxCount {
			foundMostCommon = true
			mostCommon = reactionType
			maxCount = occurances
		}
	}

	if !foundMostCommon {
		return -1
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

func (g *Game) startEmulator() {
	g.Emulator.Play(g.RomPath)
}

func (g *Game) continuoslySave() {
	c := time.Tick(1 * time.Minute)
	for range c {
		path := nesSaveFilePath(g.SavePath, g.RomPath)

		fmt.Println("Saving NES's game state to:", path)
		g.Emulator.SaveState(path)
		fmt.Println("Finished saving...")
	}
}

func nesSaveFilePath(savePath, romPath string) string {
	return filepath.Join(savePath, romPathHash(romPath), "save.dat")
}

func romPathHash(romPath string) string {
	h := md5.New()
	io.WriteString(h, romPath)

	return fmt.Sprintf("%x", h.Sum(nil))
}
