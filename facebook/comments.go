package facebook

import (
	"fmt"
	"time"

	fb "github.com/huandu/facebook"
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
	res, err := fb.Get(fmt.Sprintf("/%s/comments", id), fb.Params{
		"access_token": accessToken,
	})
	if err != nil {
		return nil, err
	}

	rawComments := res["data"].([]interface{})
	comments := make([]Comment, len(rawComments))

	for i, rawComment := range rawComments {
		coerced := rawComment.(map[string]interface{})

		created, err := time.Parse(util.ISO8601, coerced["created_time"].(string))
		if err != nil {
			return nil, err
		}

		authorInfo := coerced["from"].(map[string]interface{})

		comments[i] = Comment{
			Id:         coerced["id"].(string),
			Created:    created,
			AuthorId:   authorInfo["id"].(string),
			AuthorName: authorInfo["name"].(string),
			Message:    coerced["message"].(string),
		}
	}

	return comments, nil
}
