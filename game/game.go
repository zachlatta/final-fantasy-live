package game

import (
	"encoding/json"
	"fmt"
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
	GameSavePath = "game.json"

	actionInterval   = 10
	pollInterval     = 2 * time.Second
	inactivityCutoff = 1 * time.Minute

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
	Video       facebook.LiveVideo `json:"-"`
	RomPath     string
	SavePath    string
	AccessToken string
	Emulator    *emulator.Emulator `json:"-"`
	Obs         obs.Obs

	startTime time.Time `json: "startTime"`

	// Key is user ID
	reactions         map[string]facebook.Reaction
	lastUserReactions map[string]time.Time

	buttonsToPress chan int

	comments        chan facebook.Comment
	lastCommentTime time.Time
}

func NewFromSave(save Save, vid facebook.LiveVideo, accessToken string) (Game, error) {
	romPath := save.RomPath
	savePath := save.SavePath

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

		reactions:         save.PastReactions,
		lastUserReactions: save.LastUserReactions,

		buttonsToPress: make(chan int),

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}, nil
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

		reactions:         map[string]facebook.Reaction{},
		lastUserReactions: map[string]time.Time{},

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
	epath := nesSaveFilePath(g.SavePath, g.RomPath)
	err := g.Emulator.SaveState(epath)
	if err != nil {
		return err
	}

	save := Save{
		PastReactions:     g.reactions,
		LastUserReactions: g.lastUserReactions,
		RomPath:           g.RomPath,
		SavePath:          g.SavePath,
	}

	fpath := filepath.Join(g.SavePath, util.MD5HashString(g.RomPath), GameSavePath)

	var f *os.File

	// Check if file doesn't exist
	if ok, err := util.FileExists(fpath); !ok || err != nil {
		f, err = os.Create(fpath)
	} else {
		f, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	}

	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(save)

	return err
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

		// Update g.lastUserReactions
		for _, reaction := range reactions {
			lastReaction := g.reactions[reaction.AuthorId]

			if lastReaction != reaction {
				g.lastUserReactions[reaction.AuthorId] = time.Now()
			}
		}

		// Update g.reactions
		reactionMap := map[string]facebook.Reaction{}
		for _, reaction := range reactions {
			reactionMap[reaction.AuthorId] = reaction
		}
		g.reactions = reactionMap

		// Update vote breakdown
		buttonVoteMap := map[int]int{}
		for reactionType, count := range g.reactionCounts(true) {
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
		g.Obs.UpdateActivePlayers(len(g.activePlayers()))

		timer -= 1

		if timer == 0 {
			timer = actionInterval

			mostCommonReact := mostCommonReact(g.reactionCounts(true))
			if mostCommonReact == -1 {
				fmt.Println("No reactions. Skipping button press.")
				continue
			}

			g.buttonsToPress <- reactionToButton(mostCommonReact)
		}
	}
}

func (g Game) activePlayers() map[string]struct{} {
	activeIds := map[string]struct{}{}

	for userId, lastReactionTime := range g.lastUserReactions {
		// cutoff = current time - inactivity cutoff
		cutoff := time.Now().Add(-inactivityCutoff)

		if lastReactionTime.After(cutoff) {
			activeIds[userId] = struct{}{}
		}
	}

	return activeIds
}

func (g Game) reactionCounts(onlyIncludeActive bool) map[facebook.ReactionType]int {
	countMap := map[facebook.ReactionType]int{}
	activeUserIds := g.activePlayers()

	for userId, reaction := range g.reactions {
		if onlyIncludeActive {
			if _, present := activeUserIds[userId]; present == true {
				countMap[reaction.Type] += 1
			}
		} else {
			countMap[reaction.Type] += 1
		}
	}

	return countMap
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
		g.Save()
		fmt.Println("Finished saving...")
	}
}

func nesSaveFilePath(savePath, romPath string) string {
	return filepath.Join(savePath, util.MD5HashString(romPath), "save.dat")
}

type Save struct {
	PastReactions     map[string]facebook.Reaction `json:"past_reactions"`
	LastUserReactions map[string]time.Time         `json:"last_user_reactions"`
	RomPath           string                       `json:"rom_path"`
	SavePath          string                       `json:"save_path"`
}
