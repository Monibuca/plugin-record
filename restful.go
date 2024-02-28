package record

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		if f, err := time.ParseDuration(fragment); err == nil {
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

func (conf *RecordConfig) API_recordfile_delete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	path := query.Get("path")
	err := os.Remove(path)
	if err != nil {
		plugin.Error("修改文件时出错", zap.Error(err))
		util.ReturnError(1, "删除文件时出错", w, r)
		return
	}
	util.ReturnOK(w, r)
}

func (conf *RecordConfig) API_recordfile_modify(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	path := query.Get("path")
	newName := query.Get("newName")
	dirPath := filepath.Dir(path)
	err := os.Rename(path, dirPath+"/"+newName)
	if err != nil {
		plugin.Error("修改文件时出错", zap.Error(err))
		util.ReturnError(1, "修改文件时出错", w, r)
		return
	}
	util.ReturnOK(w, r)
}

func (conf *RecordConfig) API_list_page(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	t := query.Get("type")
	pageSize := query.Get("pageSize")
	pageNum := query.Get("pageNum")
	streamPath := query.Get("streamPath") //搜索条件
	var files []*VideoFileInfo
	var outFiles []*VideoFileInfo
	var err error
	var totalPageCount int = 1
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
	if streamPath != "" {
		for _, file := range files {
			if strings.Contains(file.Path, streamPath) {
				outFiles = append(outFiles, file)
			}
		}
	} else {
		outFiles = files
	}

	totalCount := len(outFiles) //总条数
	if pageSize != "" && pageNum != "" {
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil {
			http.Error(w, "pageSize parameter error", http.StatusBadRequest)
			return
		}
		pageNumInt, err := strconv.Atoi(pageNum)
		if err != nil {
			http.Error(w, "pageNum parameter error", http.StatusBadRequest)
			return
		}
		if pageSizeInt > 0 {
			if pageNumInt > 0 {
				totalPageCount = (totalCount / pageSizeInt) + 1 //总页数
				remainCount := totalCount % pageSizeInt         //分页后剩余条数
				if remainCount == 0 && totalPageCount != 1 {
					totalPageCount = totalCount / pageSizeInt
				}
				if pageNumInt > totalPageCount {
					http.Error(w, "pageSize parameter error", http.StatusBadRequest)
					return
				}
				startIndex := (pageNumInt - 1) * pageSizeInt //开始索引
				endIndex := startIndex + pageSizeInt
				if endIndex > totalCount {
					endIndex = totalCount
				}
				outFiles = outFiles[startIndex:endIndex]
			} else {
				http.Error(w, "pageNum parameter error", http.StatusBadRequest)
				return
			}
		}
	}

	if err == nil {
		//var bytes []byte
		//if bytes, err = json.Marshal(&struct {
		//	Files          []*VideoFileInfo
		//	TotalCount     int
		//	TotalPageCount int
		//}{
		//	Files:          outFiles,
		//	TotalCount:     totalCount,
		//	TotalPageCount: totalPageCount,
		//}); err == nil {
		//	w.Write(bytes)
		//}
		util.ReturnValue(&struct {
			Files          []*VideoFileInfo `json:"files"`
			TotalCount     int              `json:"totalCount"`
			TotalPageCount int              `json:"totalPageCount"`
		}{
			Files:          outFiles,
			TotalCount:     totalCount,
			TotalPageCount: totalPageCount,
		}, w, r)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (conf *RecordConfig) API_list_recording_page(w http.ResponseWriter, r *http.Request) {
	util.ReturnFetchValue(func() (outdata any) {
		var recordings []any
		conf.recordings.Range(func(key, value any) bool {
			recordings = append(recordings, value)
			return true
		})
		query := r.URL.Query()
		pageSize := query.Get("pageSize")
		pageNum := query.Get("pageNum")
		ID := query.Get("ID") //搜索条件
		var outRecordings []any
		var totalPageCount int = 1
		if ID != "" {
			for _, record := range recordings {
				if strings.Contains(record.(IRecorder).GetRecorder().ID, ID) {
					outRecordings = append(outRecordings, record)
				}
			}
		} else {
			outRecordings = recordings
		}

		totalCount := len(outRecordings) //总条数
		if pageSize != "" && pageNum != "" {
			pageSizeInt, err := strconv.Atoi(pageSize)
			if err != nil {
				http.Error(w, "pageSize parameter error", http.StatusBadRequest)
				return
			}
			pageNumInt, err := strconv.Atoi(pageNum)
			if err != nil {
				http.Error(w, "pageNum parameter error", http.StatusBadRequest)
				return
			}
			if pageSizeInt > 0 {
				if pageNumInt > 0 {
					totalPageCount = (totalCount / pageSizeInt) + 1 //总页数
					remainCount := totalCount % pageSizeInt         //分页后剩余条数
					if remainCount == 0 && totalPageCount != 1 {
						totalPageCount = totalCount / pageSizeInt
					}
					if pageNumInt > totalPageCount {
						http.Error(w, "pageSize parameter error", http.StatusBadRequest)
						return
					}
					startIndex := (pageNumInt - 1) * pageSizeInt //开始索引
					endIndex := startIndex + pageSizeInt
					if endIndex > totalCount {
						endIndex = totalCount
					}
					outRecordings = outRecordings[startIndex:endIndex]
				} else {
					http.Error(w, "pageNum parameter error", http.StatusBadRequest)
					return
				}
			}
		}
		outdata = &struct {
			Files          []any `json:"files"`
			TotalCount     int   `json:"totalCount"`
			TotalPageCount int   `json:"totalPageCount"`
		}{
			Files:          outRecordings,
			TotalCount:     totalCount,
			TotalPageCount: totalPageCount,
		}
		return
	}, w, r)
}
