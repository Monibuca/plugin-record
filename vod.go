package record

import (
	"net/http"
)

func ext(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func (conf *RecordConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch ext(r.URL.Path) {
	case ".flv":
		conf.Flv.ServeHTTP(w, r)
	case ".mp4":
		conf.Mp4.ServeHTTP(w, r)
	case ".m3u8", ".ts":
		conf.Hls.ServeHTTP(w, r)
	case ".h264", ".h265":
		conf.Raw.ServeHTTP(w, r)
	}
}
