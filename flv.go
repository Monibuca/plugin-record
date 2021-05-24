package record

import (
	"io"
	"os"
	"path/filepath"

	. "github.com/Monibuca/engine/v2"
	"github.com/Monibuca/engine/v2/avformat"
	"github.com/Monibuca/engine/v2/util"
)

func getDuration(file *os.File) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := avformat.ReadFLVTag(file)
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
	os.MkdirAll(filepath.Dir(filePath), 0775)
	file, err := os.OpenFile(filePath, flag, 0775)
	if err != nil {
		return err
	}
	p := Subscriber{OnData: func(packet *avformat.SendPacket) error {
		return avformat.WriteFLVTag(file, packet)
	}}
	p.ID = filePath
	p.Type = "FlvRecord"
	if append {
		p.OffsetTime = getDuration(file)
		file.Seek(0, io.SeekEnd)
	} else {
		_, err = file.Write(avformat.FLVHeader)
	}
	if err == nil {
		recordings.Store(filePath, &p)
		go func() {
			p.Subscribe(streamPath)
			file.Close()
		}()
	} else {
		file.Close()
	}
	return err
}
