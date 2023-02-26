package record

import (
	"path/filepath"
	"strconv"
	"time"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
)

type RawRecorder struct {
	Recorder
}

func (r *RawRecorder) Start(streamPath string) error {
	r.Record = &RecordPluginConfig.Raw
	r.ID = streamPath + "/raw"
	if _, ok := RecordPluginConfig.recordings.Load(r.ID); ok {
		return ErrRecordExist
	}
	return plugin.Subscribe(streamPath, r)
}

func (r *RawRecorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case *RawRecorder:
		go r.start()
	case *track.Video:
		if r.Ext == "." {
			if v.CodecID == codec.CodecID_H264 {
				r.Ext = ".h264"
			} else {
				r.Ext = ".h265"
			}
		}
	case VideoFrame:
		if r.Fragment != 0 && r.newFile {
			r.newFile = false
			r.Close()
			if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
				r.SetIO(file)
			}
		}
		v.WriteAnnexBTo(r)
	}
}
