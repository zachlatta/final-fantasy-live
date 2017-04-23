package facebook

import (
	fb "github.com/huandu/facebook"
)

func authedSession(accessToken string) *fb.Session {
	session := fb.Session{}
	session.SetAccessToken(accessToken)

	return &session
}

func getAllPaginated(session *fb.Session, path string, params fb.Params) ([]fb.Result, error) {
	res, err := session.Get(path, params)
	if err != nil {
		return nil, err
	}

	paging, err := res.Paging(session)
	if err != nil {
		return nil, err
	}

	noMore := !paging.HasNext()

	for !noMore {
		noMore, err = paging.Next()
		if err != nil {
			return nil, err
		}
	}

	return paging.Data(), nil
}
