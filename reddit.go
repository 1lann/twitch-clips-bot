package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/api/youtube/v3"

	"github.com/turnage/graw"
	"github.com/turnage/redditproto"
)

type twitchClipsBot struct {
	eng graw.Engine
}

type clip struct {
	video       io.ReadCloser
	username    string
	displayName string
	videoURL    string
}

var monitoredSubreddits = []string{"leagueoflegends", "1lann"}

var yt *youtube.Service

var linkRe = regexp.MustCompile(`https:\/\/clips\.twitch\.tv\/[^\s\]\)\.]+`)
var clipRe = regexp.MustCompile(`var clipInfo = {.+\sbroadcaster_login: "([^"]+)",.+\sbroadcaster_display_name: "([^"]+)",.+\sclip_video_url: "([^"]+)",`)

func (b *twitchClipsBot) SetUp() error {
	b.eng = graw.GetEngine(b)
	return nil
}

func parsePost(post *redditproto.Link) (string, string, bool) {
	if post.GetIsSelf() {
		results := linkRe.FindAllString(post.GetSelftext(), 1)
		if len(results) == 0 {
			return "", "", false
		}

		return post.GetTitle(), results[0], true
	}

	url := post.GetUrl()
	if !strings.HasPrefix(url, "https://clips.twitch.tv/") {
		return "", "", false
	}

	return post.GetTitle(), url, true
}

func getClipMetadata(url string) (*clip, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("metadata: " + resp.Status)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	noLines := strings.Replace(string(data), "\n", "", -1)
	result := clipRe.FindAllStringSubmatch(noLines, 1)
	if len(result) == 0 {
		return nil, errors.New("metadata: failed to parse metadata")
	}

	return &clip{
		username:    strings.Replace(result[0][1], "\\", "", -1),
		displayName: strings.Replace(result[0][2], "\\", "", -1),
		videoURL:    strings.Replace(result[0][3], "\\", "", -1),
	}, nil
}

func getClip(url string) (*clip, error) {
	c, err := getClipMetadata(url)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(c.videoURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("clip: " + resp.Status)
	}

	c.video = resp.Body

	return c, nil
}

func makeVideoDescription(c *clip, url string, post *redditproto.Link) string {
	return "Automated mirror of: " + url +
		"\nRecorded from " + c.displayName +
		"'s Twitch channel: https://wwww.twitch.tv/" + c.username +
		"\nDiscuss on Reddit: https://wwww.reddit.com" + post.GetPermalink() +
		"\n\nThis video was uploaded automatically by a bot. " +
		"This video is not owned by this automated channel or the creator " +
		"of this automated channel. It is owned by the respective streamer whose " +
		"channel can be found above." +
		"\n\nThe source code of this bot can be found on GitHub: " +
		"https://github.com/1lann/twitch-clips-bot" +
		"\n\nEmail bot@chuie.io if you have any inquiries or issues. If you're " +
		"the broadcaster of this clip, I'll be more than happy to take down " +
		"this video if you request me to do so."
}

func makePostReply(id string) string {
	return "[YouTube mirror](https://www.youtube.com/watch?v=" + id + ")\n\n" +
		"^(I'm a bot. My creator is /u/1lann. Email me at bot@chuie.io. " +
		"If you're the broadcaster of this clip, or a moderator of this " +
		"subreddit, I'll be more than happy to take down this video if you " +
		"request me to do so.) " +
		"[^(Source on GitHub)](https://github.com/1lann/twitch-clips-bot)^."
}

func (b *twitchClipsBot) Post(post *redditproto.Link) {
	title, url, ok := parsePost(post)
	if !ok {
		return
	}

	log.Println("attempting to upload video from " + post.GetPermalink())

	c, err := getClip(url)
	if err != nil {
		log.Println("get clip: "+url+":", err)
		return
	}

	defer c.video.Close()

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: makeVideoDescription(c, url, post),
			CategoryId:  "20",
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "unlisted"},
	}

	call := yt.Videos.Insert("snippet,status", upload)

	resp, err := call.Media(c.video).Do()
	if err != nil {
		log.Println("youtube:", err)
		return
	}

	log.Println("uploaded video with title \"" + title + "\" and ID \"" +
		resp.Id + "\"")

	err = b.eng.Reply(post.GetName(), makePostReply(resp.Id))
	if err != nil {
		log.Println("reddit comment:", err)
		return
	}

	log.Println("successfully made comment")
}
