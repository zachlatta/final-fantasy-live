package facebook

import (
	"fmt"
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
	session := authedSession(accessToken)
	rawReacts, err := getAllPaginated(session, fmt.Sprintf("/%s/reactions", id), nil)
	if err != nil {
		return nil, err
	}

	reacts := make([]Reaction, len(rawReacts))

	for i, rawReact := range rawReacts {
		reacts[i] = Reaction{
			AuthorId:   rawReact["id"].(string),
			AuthorName: rawReact["name"].(string),
			Type:       reactionTypeForName(rawReact["type"].(string)),
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
