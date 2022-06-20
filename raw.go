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

func (r *RawRecorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case *RawRecorder:
		go r.Start()
	case *track.Video:
		if r.Ext == "." {
			if v.CodecID == codec.CodecID_H264 {
				r.Ext = ".h264"
			} else {
				r.Ext = ".h265"
			}
		}
	case VideoDeConf:
		annexB := v.GetAnnexB()
		annexB.WriteTo(r)
	case *VideoFrame:
		if r.Fragment != 0 && r.newFile {
			r.newFile = false
			r.Close()
			if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
				r.SetIO(file)
				if r.Video.Track != nil {
					annexB := VideoDeConf(r.Video.Track.DecoderConfiguration).GetAnnexB()
					annexB.WriteTo(r)
				}
			}
		}
		annexB := v.GetAnnexB()
		annexB.WriteTo(r)
	}
}
