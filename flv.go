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

func (r *FLVRecorder) Start(streamPath string) (err error) {
	r.Record = &RecordPluginConfig.Flv
	r.ID = streamPath + "/flv"
	return plugin.Subscribe(streamPath, r)
}

func (r *FLVRecorder) start() {
	RecordPluginConfig.recordings.Store(r.ID, r)
	r.PlayFLV()
	RecordPluginConfig.recordings.Delete(r.ID)
	r.Close()
}

func (r *FLVRecorder) OnEvent(event any) {
	switch v := event.(type) {
	case ISubscriber:
		filename := strconv.FormatInt(time.Now().Unix(), 10) + r.Ext
		if r.Fragment == 0 {
			filename = r.Stream.Path + r.Ext
		} else {
			filename = filepath.Join(r.Stream.Path, filename)
		}
		if file, err := r.CreateFileFn(filename, r.append); err == nil {
			r.SetIO(file)
		}
		// 写入文件头
		if !r.append {
			r.Write(codec.FLVHeader)
		}
		go r.start()
	case FLVFrame:
		if ts := r.Video.Frame.AbsTime - r.SkipTS; r.Video.Frame.IFrame && int64(ts-r.FirstAbsTS) >= int64(r.Fragment*1000) {
			r.FirstAbsTS = ts
			r.newFile = true
		}
		if r.Fragment != 0 && r.newFile {
			r.newFile = false
			r.Close()
			if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
				r.SetIO(file)
				r.Write(codec.FLVHeader)
				if r.Video.Track != nil {
					var flvTag FLVFrame
					append(flvTag, r.Video.Track.DecoderConfiguration.FLV...).WriteTo(r)
				}
				if r.Audio.Track != nil && r.Audio.Track.CodecID == codec.CodecID_AAC {
					var flvTag FLVFrame
					append(flvTag, r.Audio.Track.DecoderConfiguration.FLV...).WriteTo(r)
				}
			}
		}
		if _, err := v.WriteTo(r); err != nil {
			r.Stop()
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
