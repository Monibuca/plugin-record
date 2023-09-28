package record

import (
	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
)

type RawRecorder struct {
	Recorder
	IsAudio bool
}

func NewRawRecorder() (r *RawRecorder) {
	r = &RawRecorder{}
	r.Record = RecordPluginConfig.Raw
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
func (r *RawRecorder) Close() (err error) {
	if r.File != nil {
		err = r.File.Close()
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
