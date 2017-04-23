package facebook

import (
	"fmt"
	"time"

	"github.com/zachlatta/nostalgic-rewind/util"
)

type Comment struct {
	Id         string
	Created    time.Time
	AuthorName string
	AuthorId   string
	Message    string
}

func Comments(id, accessToken string) ([]Comment, error) {
	session := authedSession(accessToken)

	rawComments, err := getAllPaginated(session, fmt.Sprintf("/%s/comments", id), nil)
	if err != nil {
		return nil, err
	}

	comments := make([]Comment, len(rawComments))

	for i, rawComment := range rawComments {
		created, err := time.Parse(util.ISO8601, rawComment["created_time"].(string))
		if err != nil {
			return nil, err
		}

		authorInfo := rawComment["from"].(map[string]interface{})

		comments[i] = Comment{
			Id:         rawComment["id"].(string),
			Created:    created,
			AuthorId:   authorInfo["id"].(string),
			AuthorName: authorInfo["name"].(string),
			Message:    rawComment["message"].(string),
		}
	}

	return comments, nil
}
