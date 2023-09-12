package record

import (
	"bufio"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
)

type IRecorder interface {
	ISubscriber
	GetRecorder() *Recorder
	Start(streamPath string) error
	io.Closer
	CreateFile() (FileWr, error)
}

type Recorder struct {
	Subscriber
	SkipTS   uint32
	Record   `json:"-" yaml:"-"`
	File     FileWr `json:"-" yaml:"-"`
	FileName string // 自定义文件名，分段录像无效
	filePath string // 文件路径
	append   bool   // 是否追加模式
}

func (r *Recorder) GetRecorder() *Recorder {
	return r
}

func (r *Recorder) CreateFile() (f FileWr, err error) {
	r.filePath = r.getFileName(r.Stream.Path) + r.Ext
	f, err = r.CreateFileFn(r.filePath, r.append)
	logFields := []zap.Field{zap.String("path", r.filePath)}
	if fw, ok := f.(*FileWriter); ok && r.Config != nil {
		if r.Config.WriteBufferSize > 0 {
			logFields = append(logFields, zap.Int("bufferSize", r.Config.WriteBufferSize))
			fw.bufw = bufio.NewWriterSize(fw.Writer, r.Config.WriteBufferSize)
			fw.Writer = fw.bufw
		}
	}
	if err == nil {
		r.Info("create file", logFields...)
	} else {
		logFields = append(logFields, zap.Error(err))
		r.Error("create file", logFields...)
	}
	return
}

func (r *Recorder) getFileName(streamPath string) (filename string) {
	filename = streamPath
	if r.Fragment == 0 {
		if r.FileName != "" {
			filename = filepath.Join(filename, r.FileName)
		}
	} else {
		filename = filepath.Join(filename, strconv.FormatInt(time.Now().Unix(), 10))
	}
	return
}

func (r *Recorder) start(re IRecorder, streamPath string, subType byte) (err error) {
	err = plugin.Subscribe(streamPath, re)
	if err == nil {
		if _, loaded := RecordPluginConfig.recordings.LoadOrStore(r.ID, re); loaded {
			return ErrRecordExist
		}
		r.Closer = re
		go func() {
			r.PlayBlock(subType)
			RecordPluginConfig.recordings.Delete(r.ID)
		}()
	}
	return
}

func (r *Recorder) cut(absTime uint32) {
	if ts := absTime - r.SkipTS; time.Duration(ts)*time.Millisecond >= r.Fragment {
		r.SkipTS = absTime
		r.Close()
		if file, err := r.Spesific.(IRecorder).CreateFile(); err == nil {
			r.File = file
			r.Spesific.OnEvent(file)
		} else {
			r.Stop(zap.Error(err))
		}
	}
}

// func (r *Recorder) Stop(reason ...zap.Field) {
// 	r.Close()
// 	r.Subscriber.Stop(reason...)
// }

func (r *Recorder) OnEvent(event any) {
	switch v := event.(type) {
	case IRecorder:
		if file, err := r.Spesific.(IRecorder).CreateFile(); err == nil {
			r.File = file
			r.Spesific.OnEvent(file)
		} else {
			r.Stop(zap.Error(err))
		}
	case AudioFrame:
		// 纯音频流的情况下需要切割文件
		if r.Fragment > 0 && r.VideoReader == nil {
			r.cut(v.AbsTime)
		}
	case VideoFrame:
		if r.Fragment > 0 && v.IFrame {
			r.cut(v.AbsTime)
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
