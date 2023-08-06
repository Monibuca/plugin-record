package record

import (
	"io"
	"net"
	"os"
	"time"

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
}

func NewFLVRecorder() (r *FLVRecorder) {
	r = &FLVRecorder{}
	r.Record = RecordPluginConfig.Flv
	return r
}

func (r *FLVRecorder) Start(streamPath string) (err error) {
	r.ID = streamPath + "/flv"
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
			if file, err := r.createFile(); err == nil {
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

func (r *FLVRecorder) Close() error {
	if r.File != nil {
		if !r.append {
			go r.writeMetaData(r.File, r.duration)
		} else {
			return r.File.Close()
		}
	}
	return nil
}
