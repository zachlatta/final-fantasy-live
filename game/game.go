package game

import (
	"fmt"
	"os"
	"time"

	"github.com/zachlatta/nostalgic-rewind/facebook"
)

const (
	reactPollInterval = 2 * time.Second
)

type Game struct {
	Video       facebook.LiveVideo
	GameName    string
	AccessToken string

	comments        chan facebook.Comment
	lastCommentTime time.Time
}

func New(vid facebook.LiveVideo, gameName string, accessToken string) Game {
	return Game{
		Video:       vid,
		GameName:    gameName,
		AccessToken: accessToken,

		comments:        make(chan facebook.Comment),
		lastCommentTime: time.Now(),
	}
}

func (g Game) Start() {
	fmt.Println("Stream created!")
	fmt.Println("ID:", g.Video.Id)
	fmt.Println("Direct your stream to:", g.Video.StreamUrl)

	go g.listenForComments()

	for comment := range g.comments {
		fmt.Println("New comment:", comment.Message)
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
