package crawler

import "github.com/hanke0/subtitles-robot/video"

type Crawler interface {
	FindMovie(v *video.Video)
	FindSeries(v *video.Video)
	FindAnime(v *video.Video)
}
