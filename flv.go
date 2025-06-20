package record

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/util"
)

type FLVRecorder struct {
	Recorder
	filepositions []uint64
	times         []float64
	Offset        int64
	duration      int64
	timer         *time.Timer
	stopCh        chan struct{}
	mu            sync.Mutex
	RecordMode
}

func (r *FLVRecorder) SetId(streamPath string) {
	r.ID = fmt.Sprintf("%s/flv/%s", streamPath, r.GetRecordModeString(r.RecordMode))
}

func (r *FLVRecorder) GetRecordModeString(mode RecordMode) string {
	switch mode {
	case EventMode:
		return "eventmode"
	case OrdinaryMode:
		return "ordinarymode"
	default:
		return ""
	}
}

// Goroutine 等待定时器停止录像
func (r *FLVRecorder) waitForStop(streamPath string) {
	select {
	case <-r.timer.C: // 定时器到期
		r.StopTimerRecord(zap.String("reason", "timer expired"))
	case <-r.stopCh: // 手动停止
		return
	}
}

// 停止定时录像
func (r *FLVRecorder) StopTimerRecord(reason ...zapcore.Field) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 停止录像
	r.Stop(reason...)

	// 关闭 stop 通道，停止 Goroutine
	close(r.stopCh)
}

// 重置定时器
func (r *FLVRecorder) resetTimer(timeout time.Duration) {
	if r.timer != nil {
		r.Info("事件录像", zap.String("timeout seconeds is reset to", fmt.Sprintf("%.0f", timeout.Seconds())))
		r.timer.Reset(timeout)
	} else {
		r.Info("事件录像", zap.String("timeout seconeds is first set to", fmt.Sprintf("%.0f", timeout.Seconds())))
		r.timer = time.NewTimer(timeout)
	}
}

func (r *FLVRecorder) StartWithDynamicTimeout(streamPath, fileName string, timeout time.Duration) error {
	// 启动录像
	if err := r.StartWithFileName(streamPath, fileName); err != nil {
		return err
	}

	// 创建定时器
	r.resetTimer(timeout)

	// 启动 Goroutine 监听定时器
	go r.waitForStop(streamPath)

	return nil
}

func (r *FLVRecorder) UpdateTimeout(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 停止旧的定时器并重置
	r.resetTimer(timeout)
}

func NewFLVRecorder(mode RecordMode) (r *FLVRecorder) {
	r = &FLVRecorder{
		stopCh:     make(chan struct{}),
		RecordMode: mode,
	}
	r.Record = RecordPluginConfig.Flv
	return r
}

func (r *FLVRecorder) Start(streamPath string) (err error) {
	r.ID = fmt.Sprintf("%s/flv/%s", streamPath, r.GetRecordModeString(r.RecordMode))
	return r.start(r, streamPath, SUBTYPE_FLV)
}

func (r *FLVRecorder) StartWithFileName(streamPath string, fileName string) error {
	r.ID = fmt.Sprintf("%s/flv/%s", streamPath, fileName)
	return r.start(r, streamPath, SUBTYPE_FLV)
}

func (r *FLVRecorder) writeMetaData(file FileWr, duration int64) {
	defer file.Close()
	at, vt := r.Audio, r.Video
	hasAudio, hasVideo := at != nil, vt != nil
	var amf util.AMF
	metaData := util.EcmaArray{
		"MetaDataCreator": "m7s " + Engine.Version,
		"hasVideo":        hasVideo,
		"hasAudio":        hasAudio,
		"hasMatadata":     true,
		"canSeekToEnd":    true,
		"duration":        float64(duration) / 1000,
		"hasKeyFrames":    len(r.filepositions) > 0,
		"filesize":        0,
	}
	var flags byte
	if hasAudio {
		flags |= (1 << 2)
		metaData["audiocodecid"] = int(at.CodecID)
		metaData["audiosamplerate"] = at.SampleRate
		metaData["audiosamplesize"] = at.SampleSize
		metaData["stereo"] = at.Channels == 2
	}
	if hasVideo {
		flags |= 1
		metaData["videocodecid"] = int(vt.CodecID)
		metaData["width"] = vt.SPSInfo.Width
		metaData["height"] = vt.SPSInfo.Height
		metaData["framerate"] = vt.FPS
		metaData["videodatarate"] = vt.BPS
		metaData["keyframes"] = map[string]any{
			"filepositions": r.filepositions,
			"times":         r.times,
		}
		defer func() {
			r.filepositions = []uint64{0}
			r.times = []float64{0}
		}()
	}
	amf.Marshals("onMetaData", metaData)
	offset := amf.Len() + len(codec.FLVHeader) + 15
	if keyframesCount := len(r.filepositions); keyframesCount > 0 {
		metaData["filesize"] = uint64(offset) + r.filepositions[keyframesCount-1]
		for i := range r.filepositions {
			r.filepositions[i] += uint64(offset)
		}
		metaData["keyframes"] = map[string]any{
			"filepositions": r.filepositions,
			"times":         r.times,
		}
	}

	if tempFile, err := os.CreateTemp("", "*.flv"); err != nil {
		r.Error("create temp file failed: ", zap.Error(err))
		return
	} else {
		defer func() {
			tempFile.Close()
			os.Remove(tempFile.Name())
			r.Info("writeMetaData success")
		}()
		_, err := tempFile.Write([]byte{'F', 'L', 'V', 0x01, flags, 0, 0, 0, 9, 0, 0, 0, 0})
		if err != nil {
			r.Error("", zap.Error(err))
			return
		}
		amf.Reset()
		marshals := amf.Marshals("onMetaData", metaData)
		codec.WriteFLVTag(tempFile, codec.FLV_TAG_TYPE_SCRIPT, 0, marshals)
		_, err = file.Seek(int64(len(codec.FLVHeader)), io.SeekStart)
		if err != nil {
			r.Error("writeMetaData Seek failed: ", zap.Error(err))
			return
		}
		_, err = io.Copy(tempFile, file)
		if err != nil {
			r.Error("writeMetaData Copy failed: ", zap.Error(err))
			return
		}
		tempFile.Seek(0, io.SeekStart)
		file.Seek(0, io.SeekStart)
		_, err = io.Copy(file, tempFile)
		if err != nil {
			r.Error("writeMetaData Copy failed: ", zap.Error(err))
			return
		}
	}
}

func (r *FLVRecorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case FileWr:
		// 写入文件头
		if !r.append {
			v.Write(codec.FLVHeader)
		} else {
			if _, err := v.Seek(-4, io.SeekEnd); err != nil {
				r.Error("seek file failed", zap.Error(err))
				v.Write(codec.FLVHeader)
			} else {
				tmp := make(util.Buffer, 4)
				tmp2 := tmp
				v.Read(tmp)
				tagSize := tmp.ReadUint32()
				tmp = tmp2
				v.Seek(int64(tagSize), io.SeekEnd)
				v.Read(tmp2)
				ts := tmp2.ReadUint24() | (uint32(tmp[3]) << 24)
				r.Info("append flv", zap.Uint32("last tagSize", tagSize), zap.Uint32("last ts", ts))
				if r.VideoReader != nil {
					r.VideoReader.StartTs = time.Duration(ts) * time.Millisecond
				}
				if r.AudioReader != nil {
					r.AudioReader.StartTs = time.Duration(ts) * time.Millisecond
				}
				v.Seek(0, io.SeekEnd)
			}
		}
	case VideoFrame:
		if r.VideoReader.Value.IFrame {
			//go func() { //将视频关键帧的数据存入sqlite数据库中
			//	var flvKeyfram = &FLVKeyframe{FLVFileName: r.Path + "/" + strings.ReplaceAll(r.filePath, "\\", "/"), FrameOffset: r.Offset, FrameAbstime: r.VideoReader.AbsTime}
			//	db.Create(flvKeyfram)
			//}()
			//r.Info("这是关键帧，且取到了r.filePath是" + r.Path + r.filePath)
			//r.Info("这是关键帧，且取到了r.VideoReader.AbsTime是" + strconv.FormatUint(uint64(r.VideoReader.AbsTime), 10))
			//r.Info("这是关键帧，且取到了r.Offset是" + strconv.Itoa(int(r.Offset)))
			//r.Info("这是关键帧，且取到了r.Offset是" + r.Stream.Path)
		}
	case FLVFrame:
		check := false
		var absTime uint32
		if r.VideoReader == nil {
			check = true
			absTime = r.AudioReader.AbsTime
		} else if v.IsVideo() {
			check = r.VideoReader.Value.IFrame
			absTime = r.VideoReader.AbsTime
			if check {
				r.filepositions = append(r.filepositions, uint64(r.Offset))
				r.times = append(r.times, float64(absTime)/1000)
			}
		}

		if r.duration = int64(absTime); r.Fragment > 0 && check && time.Duration(r.duration)*time.Millisecond >= r.Fragment {
			r.Close()
			r.Offset = 0
			if file, err := r.CreateFile(); err == nil {
				r.File = file
				file.Write(codec.FLVHeader)
				var dcflv net.Buffers
				if r.VideoReader != nil {
					r.VideoReader.ResetAbsTime()
					dcflv = codec.VideoAVCC2FLV(0, r.VideoReader.Track.SequenceHead)
					flv := append(dcflv, codec.VideoAVCC2FLV(0, r.VideoReader.Value.AVCC.ToBuffers()...)...)
					flv.WriteTo(file)
				}
				if r.AudioReader != nil {
					r.AudioReader.ResetAbsTime()
					if r.Audio.CodecID == codec.CodecID_AAC {
						dcflv = codec.AudioAVCC2FLV(0, r.AudioReader.Track.SequenceHead)
					}
					flv := append(dcflv, codec.AudioAVCC2FLV(0, r.AudioReader.Value.AVCC.ToBuffers()...)...)
					flv.WriteTo(file)
				}
				return
			}
		}
		if n, err := v.WriteTo(r.File); err != nil {
			r.Error("write file failed", zap.Error(err))
			r.Stop(zap.Error(err))
		} else {
			r.Offset += n
		}
	}
}

func (r *FLVRecorder) Close() (err error) {
	if r.File != nil {
		if !r.append {
			go func() {
				if r.RecordMode == OrdinaryMode {
					startTime := time.Now().Add(-time.Duration(r.duration) * time.Millisecond).Format("2006-01-02 15:04:05")
					endTime := time.Now().Format("2006-01-02 15:04:05")
					fileName := r.FileName
					if r.FileName == "" {
						fileName = strings.ReplaceAll(r.Stream.Path, "/", "-") + "-" + time.Now().Format("2006-01-02-15-04-05")
					}
					filepath := RecordPluginConfig.Flv.Path + "/" + r.Stream.Path + "/" + fileName + r.Ext //录像文件存入的完整路径（相对路径）
					eventRecord := EventRecord{StreamPath: r.Stream.Path, RecordMode: "0", BeforeDuration: "0",
						AfterDuration: fmt.Sprintf("%.0f", r.Fragment.Seconds()), CreateTime: startTime, StartTime: startTime,
						EndTime: endTime, Filepath: filepath, Filename: fileName + r.Ext, Urlpath: "record/" + strings.ReplaceAll(r.filePath, "\\", "/"), Fragment: fmt.Sprintf("%.0f", r.Fragment.Seconds()), Type: "flv"}
					err = db.Omit("id", "isDelete").Create(&eventRecord).Error
				}
			}()
			plugin.Info("====into close append false===recordid is===" + r.ID + "====record type is " + r.GetRecordModeString(r.RecordMode) + "====starttime  is " + time.Now().Add(-time.Duration(r.duration)*time.Millisecond).Format("2006-01-02 15:04:05"))
			go r.writeMetaData(r.File, r.duration)
		} else {
			plugin.Info("====into close append true===recordid is===" + r.ID + "====record type is " + r.GetRecordModeString(r.RecordMode))
			err = r.File.Close()
			if err != nil {
				r.Error("FLV File Close", zap.Error(err))
			} else {
				r.Info("FLV File Close", zap.Error(err))
				go r.UploadFile(r.Path, r.filePath)
			}
			return err
		}
	}
	return nil
}
