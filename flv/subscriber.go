package flv

import (
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
)

type Recorder struct {
	Subscriber
	Append bool
}

func (r *Recorder) OnEvent(event any) {
	switch v := event.(type) {
	case ISubscriber:
		// 写入文件头
		if !r.Append {
			r.Write(codec.FLVHeader)
		}
	case HaveFLV:
		flvTag := v.GetFLV()
		if _, err := flvTag.WriteTo(r); err != nil {
			r.Stop()
		}
	}
	r.Subscriber.OnEvent(event)
}
