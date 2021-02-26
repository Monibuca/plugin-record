package record

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/Monibuca/engine/v3"
	. "github.com/Monibuca/utils/v3"
	"github.com/Monibuca/utils/v3/codec"
)

type FlvFile struct {
	Publisher
}

func PublishFlvFile(streamPath string) error {
	flvPath := filepath.Join(config.Path, streamPath+".flv")
	os.MkdirAll(filepath.Dir(flvPath), 0755)
	if file, err := os.Open(flvPath); err == nil {
		var stream FlvFile
		if stream.Publish(streamPath) {
			stream.Type = "FlvFile"
			defer stream.Close()
			file.Seek(int64(len(codec.FLVHeader)), io.SeekStart)
			var lastTime uint32
			at := NewAudioTrack()
			vt := NewVideoTrack()
			stream.SetOriginAT(at)
			for {
				if t, timestamp, payload, err := codec.ReadFLVTag(file); err == nil {
					switch t {
					case codec.FLV_TAG_TYPE_AUDIO:
						at.Push(timestamp, payload)
					case codec.FLV_TAG_TYPE_VIDEO:
						if timestamp != 0 {
							if lastTime == 0 {
								lastTime = timestamp
							}
						}
						vt.Push(VideoPack{Timestamp: timestamp, CompositionTime: BigEndian.Uint24(payload[2:5]), Payload: payload[5:]})
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
