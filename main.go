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
	Subscribe: config.Subscribe{
		SubAudio:    true,
		SubVideo:    true,
		LiveMode:    false,
		IFrameOnly:  false,
		WaitTimeout: 10,
	},
	Flv: Record{
		Path:          "record/flv",
		Ext:           ".flv",
		GetDurationFn: getFLVDuration,
	},
	Mp4: Record{
		Path: "record/mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "record/hls",
		Ext:  ".m3u8",
	},
	Raw: Record{
		Path: "record/raw",
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
			flv.Start(v.Stream.Path)
		}
		if conf.Mp4.NeedRecord(v.Stream.Path) {
			NewMP4Recorder().Start(v.Stream.Path)
		}
		if conf.Hls.NeedRecord(v.Stream.Path) {
			var hls HLSRecorder
			hls.Start(v.Stream.Path)
		}
		if conf.Raw.NeedRecord(v.Stream.Path) {
			var raw RawRecorder
			raw.Start(v.Stream.Path)
		}
	}
}
func (conf *RecordConfig) getRecorderConfigByType(t string) (recorder *Record) {
	switch t {
	case "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	}
	return
}

func (conf *RecordConfig) API_list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	t := query.Get("type")
	var files []*VideoFileInfo
	var err error
	recorder := conf.getRecorderConfigByType(t)
	if recorder == nil {
		for _, t = range []string{"flv", "mp4", "hls", "raw"} {
			recorder = conf.getRecorderConfigByType(t)
			var fs []*VideoFileInfo
			if fs, err = recorder.Tree(recorder.Path, 0); err == nil {
				files = append(files, fs...)
			}
		}
	} else {
		files, err = recorder.Tree(recorder.Path, 0)
	}

	if err == nil {
		var bytes []byte
		if bytes, err = json.Marshal(files); err == nil {
			w.Write(bytes)
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	var err error
	switch t {
	case "":
		t = "flv"
		fallthrough
	case "flv":
		var flvRecoder FLVRecorder
		flvRecoder.append = query.Get("append") != "" && util.Exist(filePath)
		err = flvRecoder.Start(streamPath)
	case "mp4":
		err = NewMP4Recorder().Start(streamPath)
	case "hls":
		var recorder HLSRecorder
		err = recorder.Start(streamPath)
	case "raw":
		var recorder RawRecorder
		recorder.append = query.Get("append") != "" && util.Exist(filePath)
		err = recorder.Start(streamPath)
	default:
		http.Error(w, "type not supported", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(sub.GetIO().ID))
}

func (conf *RecordConfig) API_list_recording(w http.ResponseWriter, r *http.Request) {
	var recordings []any
	conf.recordings.Range(func(key, value any) bool {
		recordings = append(recordings, value)
		return true
	})
	if err := json.NewEncoder(w).Encode(recordings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (conf *RecordConfig) API_stop(w http.ResponseWriter, r *http.Request) {
	if recorder, ok := conf.recordings.Load(r.URL.Query().Get("id")); ok {
		recorder.(ISubscriber).Stop()
		w.Write([]byte("ok"))
		return
	}
	http.Error(w, "no such recorder", http.StatusBadRequest)
}

func getFLVDuration(file io.ReadSeeker) uint32 {
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
