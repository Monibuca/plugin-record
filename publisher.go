package record

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/Monibuca/engine/v2"
	"github.com/Monibuca/engine/v2/avformat"
)

type FlvFile struct {
	Publisher
}

func PublishFlvFile(streamPath string) error {
	flvPath := filepath.Join(config.Path, streamPath+".flv")
	os.MkdirAll(filepath.Dir(flvPath), 0755)
	if file, err := os.Open(flvPath); err == nil {
		stream := FlvFile{}
		if stream.Publish(streamPath) {
			stream.Type = "FlvFile"
			defer stream.Close()
			stream.UseTimestamp = true
			file.Seek(int64(len(avformat.FLVHeader)), io.SeekStart)
			var lastTime uint32
			for {
				if t, timestamp, payload, err := avformat.ReadFLVTag(file); err == nil {
					switch t {
					case avformat.FLV_TAG_TYPE_AUDIO:
						stream.PushAudio(timestamp, payload)
					case avformat.FLV_TAG_TYPE_VIDEO:
						if timestamp != 0 {
							if lastTime == 0 {
								lastTime = timestamp
							}
						}
						stream.PushVideo(timestamp, payload)
						time.Sleep(time.Duration(timestamp-lastTime) * time.Millisecond)
						lastTime = timestamp
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
