package facebook

import (
	"fmt"

	fb "github.com/huandu/facebook"
)

type ReactionType int

const (
	ReactionLike ReactionType = iota
	ReactionLove
	ReactionHaha
	ReactionWow
	ReactionSad
	ReactionAngry
	ReactionThankful
)

type Reaction struct {
	AuthorId   string
	AuthorName string
	Type       ReactionType
}

func Reactions(id, accessToken string) ([]Reaction, error) {
	res, err := fb.Get(fmt.Sprintf("/%s/reactions", id), fb.Params{
		"access_token": accessToken,
	})
	if err != nil {
		return nil, err
	}

	rawReacts := res["data"].([]interface{})
	reacts := make([]Reaction, len(rawReacts))

	for i, rawReact := range rawReacts {
		coerced := rawReact.(map[string]interface{})

		reacts[i] = Reaction{
			AuthorId:   coerced["id"].(string),
			AuthorName: coerced["name"].(string),
			Type:       reactionTypeForName(coerced["type"].(string)),
		}
	}

	return reacts, nil
}

func reactionTypeForName(reactionName string) ReactionType {
	switch reactionName {
	case "LIKE":
		return ReactionLike
	case "LOVE":
		return ReactionLove
	case "HAHA":
		return ReactionHaha
	case "WOW":
		return ReactionWow
	case "SAD":
		return ReactionSad
	case "ANGRY":
		return ReactionAngry
	case "THANKFUL":
		return ReactionThankful
	default:
		return -1
	}
}
