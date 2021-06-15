package record

import (
	"io"
	"os"
	"path/filepath"

	. "github.com/Monibuca/engine/v3"
	. "github.com/Monibuca/utils/v3"
	"github.com/Monibuca/utils/v3/codec"
)

func getDuration(file *os.File) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := codec.ReadFLVTag(file)
				if err == nil {
					return timestamp
				}
			}
		}
	}
	return 0
}
func SaveFlv(streamPath string, append bool) error {
	flag := os.O_CREATE
	if append {
		flag = flag | os.O_RDWR | os.O_APPEND
	} else {
		flag = flag | os.O_TRUNC | os.O_WRONLY
	}
	filePath := filepath.Join(config.Path, streamPath+".flv")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(filePath, flag, 0755)
	if err != nil {
		return err
	}
	// return avformat.WriteFLVTag(file, packet)
	p := Subscriber{
		ID:               filePath,
		Type:             "FlvRecord",
		ByteStreamFormat: true,
	}
	var offsetTime uint32
	if append {
		offsetTime = getDuration(file)
		file.Seek(0, io.SeekEnd)
	} else {
		_, err = file.Write(codec.FLVHeader)
	}
	if err == nil {
		recordings.Store(filePath, &p)
		if err := p.Subscribe(streamPath); err == nil {
			vt, at := p.WaitVideoTrack(), p.WaitAudioTrack()
			p.OnAudio = func(audio AudioPack) {
				if !append && at.CodecID == 10 { //AAC格式需要发送AAC头
					codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_AUDIO, 0, at.ExtraData)
				}
				codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_AUDIO, audio.Timestamp+offsetTime, audio.Payload)
				p.OnAudio = func(audio AudioPack) {
					codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_AUDIO, audio.Timestamp+offsetTime, audio.Payload)
				}
			}
			p.OnVideo = func(video VideoPack) {
				if !append {
					codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_VIDEO, 0, vt.ExtraData.Payload)
				}
				codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_VIDEO, video.Timestamp+offsetTime, video.Payload)
				p.OnVideo = func(video VideoPack) {
					codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_VIDEO, video.Timestamp+offsetTime, video.Payload)
				}
			}
			go func() {
				p.Play(at, vt)
				file.Close()
			}()
		}

	} else {
		file.Close()
	}
	return err
}
