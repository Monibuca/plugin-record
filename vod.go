package record

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/Monibuca/utils/v3"
)

func VodHandler(w http.ResponseWriter, r *http.Request) {
	CORS(w, r)
	streamPath := r.RequestURI[5:]
	filePath := filepath.Join(config.Path, streamPath)
	if file, err := os.Open(filePath); err == nil {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "video/x-flv")
		io.Copy(w, file)
	} else {
		w.WriteHeader(404)
	}
}
