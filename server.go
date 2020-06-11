package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
	userStore = make(map[string]userSession)
	faClient  *fusionauth.FusionAuthClient
)

type userSession struct {
	user         fusionauth.User
	accessToken  string
	refreshToken string
}

func setupRouter() *gin.Engine {
	host := fmt.Sprintf("http://%s:%s", FAHost, FAPort)
	baseURL, _ := url.Parse(host)
	faClient = fusionauth.NewClient(httpClient, baseURL, APIKey)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("user_session", store))

	r.LoadHTMLGlob("templates/*")

	r.GET("/", indexRoute)
	r.GET("/oauth/redirect", oauthRedirectRoute)
	r.GET("/logout", logoutRoute)
	r.GET("/authenticated", authdRoute)

	return r
}

func indexRoute(c *gin.Context) {
	userSesh := getUser(c)
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"Name": userSesh.user.FirstName, "ClientID": ClientID, "Host": FAHost, "Port": FAPort})
}

func oauthRedirectRoute(c *gin.Context) {
	token, oauthErr, err := faClient.ExchangeOAuthCodeForAccessToken(
		c.Query("code"),
		ClientID,
		ClientSecret,
		"http://localhost:8080/oauth/redirect",
	)
	if err != nil {
		log.Fatal(err)
		c.String(http.StatusInternalServerError, "There was an issue.")
	}

	log.Printf("oauthError: %+v", oauthErr)
	log.Printf("token: %+v", token.AccessToken)

	userResp, faErrors, err := faClient.RetrieveUserUsingJWT(token.AccessToken)
	if err != nil {
		log.Fatal(err)
		c.String(http.StatusInternalServerError, "There was an issue.")
	}

	log.Printf("faError: %+v", faErrors)
	log.Printf("user: %+v", userResp.User)

	session := sessions.Default(c)
	session.Set("user_id", userResp.User.Id)
	session.Save()

	userSesh := userSession{
		user:         userResp.User,
		accessToken:  token.AccessToken,
		refreshToken: token.RefreshToken,
	}

	userStore[userResp.User.Id] = userSesh

	c.Redirect(http.StatusFound, "/")
}

func logoutRoute(c *gin.Context) {
	user := getUser(c)

	session := sessions.Default(c)
	session.Clear()
	session.Save()

	faClient.Logout(true, user.refreshToken)

	c.Redirect(http.StatusFound, "/")
}

func authdRoute(c *gin.Context) {
	c.String(http.StatusOK, "You're Auth'd!")
}

func getUser(c *gin.Context) (userSesh userSession) {
	session := sessions.Default(c)
	userID, ok := session.Get("user_id").(string)

	if !ok || userID == "" {
		return userSession{
			user: fusionauth.User{},
		}
	}

	return userStore[userID]
}
