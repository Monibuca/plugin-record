package record

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

type RecordConfig struct {
	config.Subscribe
	Flv        Record
	Mp4        Record
	Hls        Record
	Raw        Record
	recordings sync.Map
}

var recordConfig = &RecordConfig{
	Flv: Record{
		Path:          "./flv",
		Ext:           ".flv",
		GetDurationFn: getDuration,
	},
	Mp4: Record{
		Path: "./mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "./hls",
		Ext:  ".m3u8",
	},
	Raw :Record{
		Path: "./raw",
		Ext:  ".", // 默认h264扩展名为.h264,h265扩展名为.h265
	},
}

var plugin = InstallPlugin(recordConfig)

func (conf *RecordConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig, config.Config:
		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Hls.Init()
		conf.Raw.Init()
	case SEpublish:
		if conf.Flv.NeedRecord(v.Stream.Path) {
			var flv FLVRecorder
			flv.Record = &conf.Flv
			plugin.Subscribe(v.Stream.Path, &flv)
		}
		if conf.Mp4.NeedRecord(v.Stream.Path) {
			mp4 := NewMP4Recorder()
			mp4.Record = &conf.Mp4
			plugin.Subscribe(v.Stream.Path, mp4)
		}
		if conf.Hls.NeedRecord(v.Stream.Path) {
			var hls HLSRecorder
			hls.Record = &conf.Hls
			plugin.Subscribe(v.Stream.Path, &hls)
		}
		if conf.Raw.NeedRecord(v.Stream.Path) {
			var raw RawRecorder
			raw.Record = &conf.Raw
			plugin.Subscribe(v.Stream.Path, &raw)
		}
	}
}

func (conf *RecordConfig) API_list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	t := query.Get("type")
	var recorder *Record
	switch t {
	case "", "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	}

	if recorder != nil {
		if files, err := recorder.Tree(recorder.Path, 0); err == nil {
			var bytes []byte
			if bytes, err = json.Marshal(files); err == nil {
				w.Write(bytes)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "type not exist", http.StatusBadRequest)
	}
}

func (conf *RecordConfig) API_start(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	streamPath := query.Get("streamPath")
	if streamPath == "" {
		http.Error(w, "no streamPath", http.StatusBadRequest)
		return
	}
	t := query.Get("type")
	var sub ISubscriber
	var filePath string
	switch t {
	case "":
		t = "flv"
		fallthrough
	case "flv":
		var flvRecoder FLVRecorder
		flvRecoder.Record = &conf.Flv
		sub = &flvRecoder
		flvRecoder.append = query.Get("append") != "" && util.Exist(filePath)
	case "mp4":
		recorder := NewMP4Recorder()
		recorder.Record = &conf.Mp4
		sub = recorder
	case "hls":
		recorder := &HLSRecorder{}
		recorder.Record = &conf.Hls
		sub = recorder
	case "raw":
		recorder := &RawRecorder{}
		recorder.Record = &conf.Raw
		recorder.append = query.Get("append") != "" && util.Exist(filePath)
		sub = recorder
	default:
		http.Error(w, "type not supported", http.StatusBadRequest)
	}
	if err := plugin.Subscribe(streamPath, sub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := streamPath + "/" + t
	sub.GetIO().ID = id
	conf.recordings.Store(id, sub)
	w.Write([]byte(id))
}

func (conf *RecordConfig) API_stop(w http.ResponseWriter, r *http.Request) {
	if recorder, ok := conf.recordings.Load(r.URL.Query().Get("id")); ok {
		recorder.(ISubscriber).Stop()
		return
	}
	http.Error(w, "no such recorder", http.StatusBadRequest)
}

func getDuration(file io.ReadSeeker) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := codec.ReadFLVTag(file)
				if err == nil {
					return timestamp
				}
			}
		}
	}
	return 0
}
