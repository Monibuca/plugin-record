package recordplugin

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/Monibuca/engine"
	"github.com/Monibuca/engine/avformat"
)

type FlvFile struct {
	InputStream
}

func PublishFlvFile(streamPath string) error {
	if file, err := os.Open(filepath.Join(config.Path, streamPath+".flv")); err == nil {
		stream := FlvFile{}
		if stream.Publish(streamPath, &stream) {
			defer stream.Close()
			stream.UseTimestamp = true
			file.Seek(int64(len(avformat.FLVHeader)), io.SeekStart)
			var lastTime uint32
			for {
				if tag, err := avformat.ReadFLVTag(file); err == nil {
					switch tag.Type {
					case avformat.FLV_TAG_TYPE_AUDIO:
						stream.PushAudio(tag)
					case avformat.FLV_TAG_TYPE_VIDEO:
						if tag.Timestamp != 0 {
							if lastTime == 0 {
								lastTime = tag.Timestamp
							}
						}
						stream.PushVideo(tag)
						time.Sleep(time.Duration(tag.Timestamp-lastTime) * time.Millisecond)
						lastTime = tag.Timestamp
					}
				} else {
					return err
				}
			}
		} else {
			return errors.New("Bad Name")
		}
	} else {
		return err
	}
}
