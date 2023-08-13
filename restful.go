package record

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/util"
)

func (conf *RecordConfig) API_list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	t := query.Get("type")
	var files []*VideoFileInfo
	var err error
	recorder := conf.getRecorderConfigByType(t)
	if recorder == nil {
		for _, t = range []string{"flv", "mp4", "fmp4", "hls", "raw", "raw_audio"} {
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
	fileName := query.Get("fileName")
	fragment := query.Get("fragment")
	if streamPath == "" {
		http.Error(w, "no streamPath", http.StatusBadRequest)
		return
	}
	t := query.Get("type")
	var id string
	var err error
	var irecorder IRecorder
	switch t {
	case "":
		t = "flv"
		fallthrough
	case "flv":
		irecorder = NewFLVRecorder()
	case "mp4":
		irecorder = NewMP4Recorder()
	case "fmp4":
		irecorder = NewFMP4Recorder()
	case "hls":
		irecorder = NewHLSRecorder()
	case "raw":
		irecorder = NewRawRecorder()
	case "raw_audio":
		irecorder = NewRawAudioRecorder()
	default:
		http.Error(w, "type not supported", http.StatusBadRequest)
		return
	}
	recorder := irecorder.GetRecorder()
	if fragment != "" {
		f, err := time.ParseDuration(fragment)
		if err != nil {
			recorder.Fragment = f
		}
	}
	recorder.FileName = fileName
	recorder.append = query.Get("append") != ""
	err = irecorder.Start(streamPath)
	id = recorder.ID
	if err != nil {
		util.ReturnError(util.APIErrorInternal, err.Error(), w, r)
		return
	}
	util.ReturnError(util.APIErrorNone, id, w, r)
}

func (conf *RecordConfig) API_list_recording(w http.ResponseWriter, r *http.Request) {
	util.ReturnFetchValue(func() (recordings []any) {
		conf.recordings.Range(func(key, value any) bool {
			recordings = append(recordings, value)
			return true
		})
		return
	}, w, r)
}

func (conf *RecordConfig) API_stop(w http.ResponseWriter, r *http.Request) {
	if recorder, ok := conf.recordings.Load(r.URL.Query().Get("id")); ok {
		recorder.(ISubscriber).Stop(zap.String("reason", "api"))
		util.ReturnOK(w, r)
		return
	}
	util.ReturnError(util.APIErrorNotFound, "no such recorder", w, r)
}
