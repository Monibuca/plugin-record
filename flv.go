package record

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
)

type FLVRecorder struct {
	Recorder
	filepositions []uint64
	times         []float64
	Offset        int64
	duration      int64
}

func (r *FLVRecorder) Start(streamPath string) (err error) {
	r.Record = &RecordPluginConfig.Flv
	r.ID = streamPath + "/flv"
	if _, ok := RecordPluginConfig.recordings.Load(r.ID); ok {
		return ErrRecordExist
	}
	return plugin.Subscribe(streamPath, r)
}

func (r *FLVRecorder) start() {
	RecordPluginConfig.recordings.Store(r.ID, r)
	r.PlayFLV()
	RecordPluginConfig.recordings.Delete(r.ID)
	if file, ok := r.Writer.(*os.File); ok {
		go r.writeMetaData(file, r.duration)
	} else {
		r.Close()
	}
}

func (r *FLVRecorder) writeMetaData(file *os.File, duration int64) {
	defer file.Close()
	at, vt := r.Audio.Track, r.Video.Track
	hasAudio, hasVideo := at != nil, vt != nil
	var amf codec.AMF
	metaData := codec.EcmaArray{
		"MetaDataCreator": "m7s " + Engine.Version,
		"hasVideo":        hasVideo,
		"hasAudio":        hasAudio,
		"hasMatadata":     true,
		"canSeekToEnd":    false,
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
		tempFile.Write([]byte{'F', 'L', 'V', 0x01, flags, 0, 0, 0, 9, 0, 0, 0, 0})
		amf.Reset()
		codec.WriteFLVTag(tempFile, codec.FLV_TAG_TYPE_SCRIPT, 0, net.Buffers{amf.Marshals("onMetaData", metaData)})
		file.Seek(int64(len(codec.FLVHeader)), io.SeekStart)
		io.Copy(tempFile, file)
		tempFile.Seek(0, io.SeekStart)
		file.Seek(0, io.SeekStart)
		io.Copy(file, tempFile)
		tempFile.Close()
		os.Remove(tempFile.Name())
	}
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
		check := false
		var absTime uint32
		if r.Video.Track == nil {
			check = true
			absTime = r.Audio.Frame.AbsTime
		} else {
			check = r.Video.Frame.IFrame
			absTime = r.Video.Frame.AbsTime
			if check {
				r.filepositions = append(r.filepositions, uint64(r.Offset))
				r.times = append(r.times, float64(absTime)/1000)
			}
		}
		if r.Fragment > 0 && check && r.duration >= int64(r.Fragment*1000) {
			r.SkipTS = absTime
			if file, ok := r.Writer.(*os.File); ok {
				go r.writeMetaData(file, r.duration)
			} else {
				r.Close()
			}
			r.Offset = 0
			if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
				r.SetIO(file)
				r.Write(codec.FLVHeader)
				if r.Video.Track != nil {
					dcflv := codec.VideoAVCC2FLV(r.Video.Track.DecoderConfiguration.AVCC, 0)
					dcflv.WriteTo(r)
				}
				if r.Audio.Track != nil && r.Audio.Track.CodecID == codec.CodecID_AAC {
					dcflv := codec.AudioAVCC2FLV(r.Audio.Track.Value.AVCC, 0)
					dcflv.WriteTo(r)
				}
				flv := codec.VideoAVCC2FLV(r.Video.Frame.AVCC, 0)
				flv.WriteTo(r)
				return
			}
		}
		if n, err := v.WriteTo(r); err != nil {
			r.Stop()
		} else {
			r.Offset += n
			r.duration = int64(absTime - r.SkipTS)
		}
	default:
		r.Subscriber.OnEvent(event)
	}
}
