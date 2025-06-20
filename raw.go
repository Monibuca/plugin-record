package record

import (
	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
	"time"
)

type RawRecorder struct {
	Recorder
	IsAudio bool
}

func (r *RawRecorder) SetId(string) {
	//TODO implement me
	panic("implement me")
}

func (r *RawRecorder) GetRecordModeString(mode RecordMode) string {
	//TODO implement me
	panic("implement me")
}

func (r *RawRecorder) StartWithDynamicTimeout(streamPath, fileName string, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (r *RawRecorder) UpdateTimeout(timeout time.Duration) {
	//TODO implement me
	panic("implement me")
}

func NewRawRecorder() (r *RawRecorder) {
	r = &RawRecorder{}
	r.Record = RecordPluginConfig.Raw
	r.Storage = RecordPluginConfig.Storage
	return r
}

func NewRawAudioRecorder() (r *RawRecorder) {
	r = &RawRecorder{IsAudio: true}
	r.Record = RecordPluginConfig.RawAudio
	return r
}

func (r *RawRecorder) Start(streamPath string) error {
	r.ID = streamPath + "/raw"
	if r.IsAudio {
		r.ID += "_audio"
	}
	return r.start(r, streamPath, SUBTYPE_RAW)
}

func (r *RawRecorder) StartWithFileName(streamPath string, fileName string) error {
	r.ID = streamPath + "/raw/" + fileName
	if r.IsAudio {
		r.ID += "_audio"
	}
	return r.start(r, streamPath, SUBTYPE_RAW)
}

func (r *RawRecorder) Close() (err error) {
	if r.File != nil {
		err = r.File.Close()
		if err != nil {
			r.Error("Raw File Close", zap.Error(err))
		} else {
			r.Info("Raw File Close", zap.Error(err))
			go r.UploadFile(r.Path, r.filePath)
		}
	}
	return
}
func (r *RawRecorder) OnEvent(event any) {
	switch v := event.(type) {
	case FileWr:
		r.SetIO(v)
	case *RawRecorder:
		r.Recorder.OnEvent(event)
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
		r.Recorder.OnEvent(event)
		if _, err := v.WriteRawTo(r); err != nil {
			r.Stop(zap.Error(err))
		}
	case VideoFrame:
		r.Recorder.OnEvent(event)
		if _, err := v.WriteAnnexBTo(r); err != nil {
			r.Stop(zap.Error(err))
		}
	default:
		r.IO.OnEvent(v)
	}
}
