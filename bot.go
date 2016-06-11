package main

import (
	"log"
	"time"

	"github.com/turnage/graw"
)

func main() {
	setupYouTube()

	for {
		log.Println("graw error:", graw.Run("agent.protobuf", &twitchClipsBot{}, monitoredSubreddits...))
		time.Sleep(time.Minute * 10)
	}
}
