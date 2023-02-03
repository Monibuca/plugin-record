package record

import (
	"encoding/json"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/util"
	"net/http"
)

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
	var id string
	var filePath string
	var err error
	switch t {
	case "":
		t = "flv"
		fallthrough
	case "flv":
		var flvRecoder FLVRecorder
		flvRecoder.IsInternal = true
		flvRecoder.append = query.Get("append") != "" && util.Exist(filePath)
		err = flvRecoder.Start(streamPath)
		id = flvRecoder.ID
	case "mp4":
		recorder := NewMP4Recorder()
		recorder.IsInternal = true
		err = recorder.Start(streamPath)
		id = recorder.ID
	case "hls":
		var recorder HLSRecorder
		recorder.IsInternal = true
		err = recorder.Start(streamPath)
		id = recorder.ID
	case "raw":
		var recorder RawRecorder
		recorder.IsInternal = true
		recorder.append = query.Get("append") != "" && util.Exist(filePath)
		err = recorder.Start(streamPath)
		id = recorder.ID
	default:
		http.Error(w, "type not supported", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(id))
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
