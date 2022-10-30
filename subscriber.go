package record

import (
	"path/filepath"
	"strconv"
	"time"

	. "m7s.live/engine/v4"
)

type Recorder struct {
	Subscriber
	*Record `json:"-"`
	newFile bool // 创建了新的文件
	append  bool // 是否追加模式
}

func (r *Recorder) start() {
	RecordPluginConfig.recordings.Store(r.ID, r)
	r.PlayRaw()
	RecordPluginConfig.recordings.Delete(r.ID)
	r.Close()
}

func (r *Recorder) OnEvent(event any) {
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
	case *VideoFrame:
		if ts := v.AbsTime - r.SkipTS; v.IFrame && int64(ts-r.FirstAbsTS) >= int64(r.Fragment*1000) {
			r.FirstAbsTS = ts
			r.newFile = true
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
