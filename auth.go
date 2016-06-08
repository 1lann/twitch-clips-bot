package main

import (
	"log"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

func setupYouTube() {
	config := &oauth2.Config{
		ClientID:     authClientID,
		ClientSecret: authClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
		RedirectURL: "http://localhost:8080/oauth2callback",
		Scopes:      []string{youtube.YoutubeUploadScope},
	}

	if len(os.Args) < 2 {
		url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
		log.Println("Authenticate yourself: " + url)
		os.Exit(0)
		return
	}

	authCode := os.Args[1]

	token, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		panic(err)
	}

	yt, err = youtube.New(config.Client(oauth2.NoContext, token))
	if err != nil {
		panic(err)
	}
}
