package record

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/Monibuca/engine/v2"
	. "github.com/Monibuca/engine/v2/util"
)

var config struct {
	Path        string
	AutoPublish bool
	AutoRecord  bool
}
var recordings sync.Map

type FlvFileInfo struct {
	Path     string
	Size     int64
	Duration uint32
}

//go:embed ui/*
//go:embed README.md
var ui embed.FS

func init() {
	InstallPlugin(&PluginConfig{
		Name:   "Record",
		Type:   PLUGIN_SUBSCRIBER,
		Config: &config,
		Run:    run,
		UIFile: &ui,
		HotConfig: map[string]func(interface{}){
			"AutoPublish": func(v interface{}) {
				config.AutoPublish = v.(bool)
			},
			"AutoRecord": func(v interface{}) {
				config.AutoRecord = v.(bool)
			},
		},
	})
}
func run() {
	OnSubscribeHooks.AddHook(onSubscribe)
	OnPublishHooks.AddHook(onPublish)
	os.MkdirAll(config.Path, 0755)
	http.HandleFunc("/record/flv/list", func(writer http.ResponseWriter, r *http.Request) {
		if files, err := tree(config.Path, 0); err == nil {
			var bytes []byte
			if bytes, err = json.Marshal(files); err == nil {
				writer.Write(bytes)
			} else {
				writer.Write([]byte("{\"err\":\"" + err.Error() + "\"}"))
			}
		} else {
			writer.Write([]byte("{\"err\":\"" + err.Error() + "\"}"))
		}
	})
	http.HandleFunc("/record/flv", func(writer http.ResponseWriter, r *http.Request) {
		if streamPath := r.URL.Query().Get("streamPath"); streamPath != "" {
			if err := SaveFlv(streamPath, r.URL.Query().Get("append") == "true"); err != nil {
				writer.Write([]byte(err.Error()))
			} else {
				writer.Write([]byte("success"))
			}
		} else {
			writer.Write([]byte("no streamPath"))
		}
	})

	http.HandleFunc("/record/flv/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if streamPath := r.URL.Query().Get("streamPath"); streamPath != "" {
			filePath := filepath.Join(config.Path, streamPath+".flv")
			if stream, ok := recordings.Load(filePath); ok {
				output := stream.(*Subscriber)
				output.Close()
				w.Write([]byte("success"))
			} else {
				w.Write([]byte("no query stream"))
			}
		} else {
			w.Write([]byte("no such stream"))
		}
	})
	http.HandleFunc("/record/flv/play", func(w http.ResponseWriter, r *http.Request) {
		if streamPath := r.URL.Query().Get("streamPath"); streamPath != "" {
			if err := PublishFlvFile(streamPath); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write([]byte("success"))
			}
		} else {
			w.Write([]byte("no streamPath"))
		}
	})
	http.HandleFunc("/record/flv/delete", func(w http.ResponseWriter, r *http.Request) {
		if streamPath := r.URL.Query().Get("streamPath"); streamPath != "" {
			filePath := filepath.Join(config.Path, streamPath+".flv")
			if Exist(filePath) {
				if err := os.Remove(filePath); err != nil {
					w.Write([]byte(err.Error()))
				} else {
					w.Write([]byte("success"))
				}
			} else {
				w.Write([]byte("no such file"))
			}
		} else {
			w.Write([]byte("no streamPath"))
		}
	})
}
func onSubscribe(s *Subscriber) {
	if config.AutoPublish {
		filePath := filepath.Join(config.Path, s.StreamPath+".flv")
		if s.Publisher == nil && Exist(filePath) {
			go PublishFlvFile(s.StreamPath)
		}
	}
}
func onPublish(p *Stream) {
	if config.AutoRecord {
		go SaveFlv(p.StreamPath, false)
	}
}

func tree(dstPath string, level int) (files []*FlvFileInfo, err error) {
	var dstF *os.File
	dstF, err = os.Open(dstPath)
	if err != nil {
		return
	}
	defer dstF.Close()
	fileInfo, err := dstF.Stat()
	if err != nil {
		return
	}
	if !fileInfo.IsDir() { //如果dstF是文件
		if path.Ext(fileInfo.Name()) == ".flv" {
			p := strings.TrimPrefix(dstPath, config.Path)
			p = strings.ReplaceAll(p, "\\", "/")
			files = append(files, &FlvFileInfo{
				Path:     strings.TrimPrefix(p, "/"),
				Size:     fileInfo.Size(),
				Duration: getDuration(dstF),
			})
		}
		return
	} else { //如果dstF是文件夹
		var dir []os.FileInfo
		dir, err = dstF.Readdir(0) //获取文件夹下各个文件或文件夹的fileInfo
		if err != nil {
			return
		}
		for _, fileInfo = range dir {
			var _files []*FlvFileInfo
			_files, err = tree(filepath.Join(dstPath, fileInfo.Name()), level+1)
			if err != nil {
				return
			}
			files = append(files, _files...)
		}
		return
	}

}
