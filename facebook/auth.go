package facebook

import (
	"fmt"
	"net/url"
	"strings"

	fb "github.com/huandu/facebook"
	"net/http"
)

const baseOauthUrl = "https://www.facebook.com/v2.9/dialog/oauth"

var requiredScopes = []string{"publish_actions", "manage_pages", "publish_pages", "user_posts", "user_videos"}
var responseTypes = []string{"code", "granted_scopes"}

// Handles every step of the login process and returns a long-lived access
// token.
func Login(appId, appSecret string) (accessToken string, err error) {
	redirectUrl := "http://localhost:6262/"

	mux := http.NewServeMux()
	srv := http.Server{Addr: ":6262", Handler: mux}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		code := params.Get("code")

		token, err := exchangeCode(appId, appSecret, code, redirectUrl)
		if err != nil {
			fmt.Fprintln(w, "Error exchanging code:", err)
			return
		}

		accessToken = token

		fmt.Fprintln(w, "Success!")

		srv.Close()
	})

	fmt.Println("Open", loginUrl(appId, redirectUrl), "in your web browser to authenticate...")
	srv.ListenAndServe()
	fmt.Println("Authenticated!")

	fmt.Println("Getting long lived token...")

	longLived, err := GetLongLivedAccessToken(appId, appSecret, accessToken)
	if err != nil {
		return "", err
	}

	fmt.Println("Done!")

	return longLived, nil
}

func loginUrl(appId, redirectUrl string) string {
	url, _ := url.Parse(baseOauthUrl)
	query := url.Query()

	query.Add("client_id", appId)
	query.Add("redirect_uri", redirectUrl)
	query.Add("scope", strings.Join(requiredScopes, ","))
	query.Add("response_type", strings.Join(responseTypes, ","))

	url.RawQuery = query.Encode()

	return url.String()
}

func exchangeCode(appId, appSecret, code, redirectUri string) (string, error) {
	res, err := fb.Get("/oauth/access_token", fb.Params{
		"client_id":     appId,
		"client_secret": appSecret,
		"code":          code,
		"redirect_uri":  redirectUri,
	})
	if err != nil {
		return "", err
	}

	token := res["access_token"].(string)

	return token, nil
}

func GetLongLivedAccessToken(appId, appSecret, accessToken string) (string, error) {
	res, err := fb.Get("/oauth/access_token", fb.Params{
		"grant_type":        "fb_exchange_token",
		"client_id":         appId,
		"client_secret":     appSecret,
		"fb_exchange_token": accessToken,
	})
	if err != nil {
		return "", err
	}

	return res["access_token"].(string), nil
}
