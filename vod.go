package record

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	p := strings.TrimPrefix(r.RequestURI, "/record/")
	switch ext(p) {
	case ".flv":
		filePath := filepath.Join(conf.Flv.Path, p)
		if file, err := os.Open(filePath); err == nil {
			w.Header().Set("Transfer-Encoding", "chunked")
			w.Header().Set("Content-Type", "video/x-flv")
			io.Copy(w, file)
		} else {
			w.WriteHeader(404)
		}
	case ".mp4":
	case ".m3u8":
	}
}
