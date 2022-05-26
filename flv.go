package record

import (
	"path/filepath"
	"strconv"
	"time"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
)

type FLVRecorder struct {
	Recorder
}

func (r *FLVRecorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case ISubscriber:
		// 写入文件头
		if !r.append {
			r.Write(codec.FLVHeader)
		}
	case HaveFLV:
		if r.Fragment != 0 {
			if r.newFile {
				r.newFile = false
				if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
					r.SetIO(file)
					r.Write(codec.FLVHeader)
					if r.Video.Track != nil {
						flvTag := VideoDeConf(r.Video.Track.DecoderConfiguration).GetFLV()
						flvTag.WriteTo(r)
					}
					if r.Audio.Track != nil && r.Audio.Track.CodecID == codec.CodecID_AAC {
						flvTag := AudioDeConf(r.Audio.Track.DecoderConfiguration).GetFLV()
						flvTag.WriteTo(r)
					}
				}
			}
		}
		flvTag := v.GetFLV()
		if _, err := flvTag.WriteTo(r); err != nil {
			r.Stop()
		}
	}
}
