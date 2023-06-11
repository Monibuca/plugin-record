package record

import (
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
)

type RawRecorder struct {
	Recorder
	IsAudio bool
}

func (r *RawRecorder) Start(streamPath string) error {
	if r.IsAudio {
		r.Record = &RecordPluginConfig.RawAudio
	} else {
		r.Record = &RecordPluginConfig.Raw
	}
	r.ID = streamPath + "/raw"
	if r.IsAudio {
		r.ID += "_audio"
	}
	if _, ok := RecordPluginConfig.recordings.Load(r.ID); ok {
		return ErrRecordExist
	}
	return plugin.Subscribe(streamPath, r)
}

func (r *RawRecorder) OnEvent(event any) {
	switch v := event.(type) {
	case *RawRecorder:
		filename := strconv.FormatInt(time.Now().Unix(), 10) + r.Ext
		if r.Fragment == 0 {
			filename = r.Stream.Path + r.Ext
		} else {
			filename = filepath.Join(r.Stream.Path, filename)
		}
		if file, err := r.CreateFileFn(filename, r.append); err == nil {
			r.SetIO(file)
		} else {
			r.Error("create file failed", zap.Error(err))
			r.Stop()
		}
		go r.start()
	case *track.Video:
		if r.IsAudio {
			break
		}
		if r.Ext == "." {
			if v.CodecID == codec.CodecID_H264 {
				r.Ext = ".h264"
			} else {
				r.Ext = ".h265"
			}
		}
		r.AddTrack(v)
	case *track.Audio:
		if !r.IsAudio {
			break
		}
		if r.Ext == "." {
			switch v.CodecID {
			case codec.CodecID_AAC:
				r.Ext = ".aac"
			case codec.CodecID_PCMA:
				r.Ext = ".pcma"
			case codec.CodecID_PCMU:
				r.Ext = ".pcmu"
			}
		}
		r.AddTrack(v)
	case AudioFrame:
		if r.Fragment > 0 {
			if r.cut(v.AbsTime); r.newFile {
				r.newFile = false
				r.Close()
				if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
					r.SetIO(file)
				}
			}
		}
		v.WriteRawTo(r)
	case VideoFrame:
		if r.Fragment > 0 && v.IFrame {
			if r.cut(v.AbsTime); r.newFile {
				r.newFile = false
				r.Close()
				if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
					r.SetIO(file)
				}
			}
		}
		v.WriteAnnexBTo(r)
	default:
		r.IO.OnEvent(v)
	}
}
