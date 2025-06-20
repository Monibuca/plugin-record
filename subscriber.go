package record

import (
	"bufio"
	"io"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
)

// 录像类型
type RecordMode int

// 使用常量块和 iota 来定义枚举值
const (
	OrdinaryMode RecordMode = iota // iota 初始值为 0，表示普通录像(连续录像)，包括自动录像和手动录像
	EventMode                      // 1，表示事件录像
)

// 判断是否有写入帧，用于解决pullonstart时拉取的流为空的情况下，生成空文件的问题
var isWrifeFrame = false

type IRecorder interface {
	ISubscriber
	GetRecorder() *Recorder
	Start(streamPath string) error
	StartWithFileName(streamPath string, fileName string) error
	io.Closer
	CreateFile() (FileWr, error)
	StartWithDynamicTimeout(streamPath, fileName string, timeout time.Duration) error
	UpdateTimeout(timeout time.Duration)
	GetRecordModeString(mode RecordMode) string
	SetId(streamPath string)
}

type Recorder struct {
	Subscriber
	Storage  StorageConfig
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
	r.SaveToDB()
	return
}

// transform 函数处理字符串并返回格式化的结果
func transform(input string) string {
	parts := strings.Split(input, "/")
	if len(parts) > 1 {
		return strings.Join(parts[1:], "-") // 返回除第一个索引外的所有部分
	}
	return input // 默认返回原始输入
}

func (r *Recorder) getFileName(streamPath string) (filename string) {
	if RecordPluginConfig.RecordPathNotShowStreamPath {
		filename = streamPath
	}
	if r.Fragment == 0 {
		if r.FileName != "" {
			filename = filepath.Join(filename, r.FileName)
		}
	} else {
		filename = filepath.Join(filename, transform(streamPath)+"_"+time.Now().Format("2006-01-02-15-04-05"))
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
	if ts := absTime - r.SkipTS; (time.Duration(ts)*time.Millisecond <= r.Fragment && r.Fragment-time.Duration(ts)*time.Millisecond <= time.Second) || time.Duration(ts)*time.Millisecond >= r.Fragment {
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

func (r *Recorder) stopByDuration(absTime uint32) {
	if ts := absTime - r.SkipTS; time.Duration(ts)*time.Millisecond >= r.Duration {
		r.Info("stop recorder by duration")
		r.SkipTS = absTime
		r.Stop()
	}
}

// func (r *Recorder) Stop(reason ...zap.Field) {
// 	r.Close()
// 	r.Subscriber.Stop(reason...)
// }

func (r *Recorder) OnEvent(event any) {
	// r.Debug("🟡->🟡->🟡 Recorder OnEvent: ", zap.String("event", reflect.TypeOf(event).String()))
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
		isWrifeFrame = true
		if v.IFrame {
			//plugin.Error("this is keyframe and absTime is " + strconv.FormatUint(uint64(v.AbsTime), 10))
			//go func() { //将视频关键帧的数据存入sqlite数据库中
			//	var flvKeyfram = &FLVKeyframe{FLVFileName: r.Path + "/" + strings.ReplaceAll(r.filePath, "\\", "/"), FrameOffset: r.VideoReader, FrameAbstime: v.AbsTime}
			//	sqlitedb.Create(flvKeyfram)
			//}()
			//r.Info("这是关键帧，且取到了r.filePath是" + r.Path + r.filePath)
			//r.Info("这是关键帧，且取到了r.VideoReader.AbsTime是" + strconv.FormatUint(uint64(v.FrameAbstime), 10))
			//r.Info("这是关键帧，且取到了r.Offset是" + strconv.Itoa(int(v.FrameOffset)))
			//r.Info("这是关键帧，且取到了r.Offset是" + r.Stream.Path)
		}
		if r.Fragment > 0 && v.IFrame {
			r.cut(v.AbsTime)
		}
		if r.Duration > 0 && v.IFrame {
			r.stopByDuration(v.AbsTime)
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
