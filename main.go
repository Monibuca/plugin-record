package record

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"unsafe"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
	"m7s.live/plugin/record/v4/flv"
	"m7s.live/plugin/record/v4/mp4"
)

type RecordConfig struct {
	Subscriber
	Flv        config.Record
	Mp4        config.Record
	Hls        config.Record
	recordings sync.Map
}

var recordConfig = &RecordConfig{
	Flv: config.Record{
		Path:          "./flv",
		Ext:           ".flv",
		GetDurationFn: getDuration,
	},
	Mp4: config.Record{
		Path: "./mp4",
		Ext:  ".mp4",
	},
	Hls: config.Record{
		Path: "./hls",
		Ext:  ".m3u8",
	},
}

var plugin = InstallPlugin(recordConfig)

func (conf *RecordConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig, config.Config:
		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Hls.Init()
	case SEpublish:
		if conf.Flv.NeedRecord(v.Stream.Path) {
			var recorder flv.Recorder
			if file, err := conf.Flv.CreateFileFn(v.Stream.Path, recorder.Append); err == nil {
				go func() {
					plugin.SubscribeBlock(v.Stream.Path, &recorder)
					file.Close()
				}()
			}
		}
		if conf.Mp4.NeedRecord(v.Stream.Path) {
			if file, err := conf.Mp4.CreateFileFn(v.Stream.Path, false); err == nil {
				recorder := mp4.NewRecorder(file)
				go func() {
					plugin.SubscribeBlock(v.Stream.Path, recorder)
					recorder.Close()
				}()
			}
		}
	}
}

func (conf *RecordConfig) API_list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	t := query.Get("type")
	var recorder *config.Record
	switch t {
	case "", "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "hls":
		recorder = &conf.Hls
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
	var recorderConf *config.Record
	var recorder ISubscriber
	var filePath string
	var point unsafe.Pointer
	var closer io.Closer
	switch t {
	case "":
		t = "flv"
		fallthrough
	case "flv":
		recorderConf = &conf.Flv
		var flvRecoder flv.Recorder
		recorder = &flvRecoder
		point = unsafe.Pointer(&flvRecoder)
		filePath = filepath.Join(recorderConf.Path, streamPath+".flv")
		flvRecoder.Append = query.Get("append") != "" && util.Exist(filePath)
		file, err := recorderConf.CreateFileFn(filePath, flvRecoder.Append)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		closer = file
	case "mp4":
		recorderConf = &conf.Mp4
		filePath = filepath.Join(recorderConf.Path, streamPath+".mp4")
		file, err := recorderConf.CreateFileFn(filePath, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mp4Recorder := mp4.NewRecorder(file)
		recorder = mp4Recorder
		point = unsafe.Pointer(mp4Recorder)
		closer = mp4Recorder
	case "hls":
		recorderConf = &conf.Hls
	}
	if err := plugin.Subscribe(streamPath, recorder); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		closer.Close()
	} else {
		conf.recordings.Store(uintptr(point), recorder)
		go func() {
			recorder.PlayBlock()
			conf.recordings.Delete(uintptr(point))
			closer.Close()
		}()
		w.Write([]byte(strconv.FormatUint(uint64(uintptr(point)), 10)))
	}
}

func (conf *RecordConfig) API_stop(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	num, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if recorder, ok := conf.recordings.Load(uintptr(num)); ok {
		recorder.(ISubscriber).Stop()
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
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
