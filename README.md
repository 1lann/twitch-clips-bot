# Twitch Clips Bot
Yet another Reddit bot...

## Requirements
You need an `agent.protobuf` ([Example](https://github.com/turnage/graw/blob/master/useragent.protobuf.template)) for the Reddit API from [graw](https://github.com/turnage/graw), and a `credentials.go` (See `credentials.go.sample` for an example) for the YouTube API.

## Features
Doesn't download anything to the hard drive! The download stream is sent
directly to YouTube! i.e. the video is uploaded as it's being downloaded, making
the process super fast (under a minute).

## License
Twitch Clips Bot is licensed under the MIT license which can be found [here](/LICENSE).
