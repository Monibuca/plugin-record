package record

import (
	"io"
	"os"
	"path"
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
	os.MkdirAll(path.Dir(filePath), 0755)
	file, err := os.OpenFile(filePath, flag, 0755)
	if err != nil {
		return err
	}
	// return avformat.WriteFLVTag(file, packet)
	p := Subscriber{
		ID:   filePath,
		Type: "FlvRecord",
	}

	if append {
		p.OffsetTime = getDuration(file)
		file.Seek(0, io.SeekEnd)
	} else {
		_, err = file.Write(codec.FLVHeader)
	}
	if err == nil {
		recordings.Store(filePath, &p)
		if err := p.Subscribe(streamPath); err == nil {
			at, vt := p.OriginAudioTrack, p.OriginVideoTrack
			tag0 := at.RtmpTag[0]
			p.OnAudio = func(audio AudioPack) {
				codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_AUDIO, audio.Timestamp, audio.ToRTMPTag(tag0))
			}
			p.OnVideo = func(video VideoPack) {
				codec.WriteFLVTag(file, codec.FLV_TAG_TYPE_VIDEO, video.Timestamp, video.ToRTMPTag())
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
