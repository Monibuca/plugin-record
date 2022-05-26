package record

import (
	"path/filepath"
	"strconv"
	"time"

	. "m7s.live/engine/v4"
)

type Recorder struct {
	Subscriber
	*Record
	newFile bool   // 创建了新的文件
	append  bool   // 是否追加模式
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
			go func() {
				r.PlayBlock()
				recordConfig.recordings.Delete(r.ID)
				r.Close()
			}()
		}
	case *VideoFrame:
		if ts := v.AbsTime; v.IFrame && int64(ts-r.Video.First.AbsTime) >= int64(r.Fragment*1000) {
			r.Video.First.AbsTime = ts
			r.newFile = true
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
