package record

import (
	_ "embed"
	"errors"
	"io"
	"sync"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

type RecordConfig struct {
	DefaultYaml
	config.Subscribe
	Flv        Record
	Mp4        Record
	Hls        Record
	Raw        Record
	recordings sync.Map
}

//go:embed default.yaml
var defaultYaml DefaultYaml
var ErrRecordExist = errors.New("recorder exist")
var RecordPluginConfig = &RecordConfig{
	DefaultYaml: defaultYaml,
	Flv: Record{
		Path:          "record/flv",
		Ext:           ".flv",
		GetDurationFn: getFLVDuration,
	},
	Mp4: Record{
		Path: "record/mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "record/hls",
		Ext:  ".m3u8",
	},
	Raw: Record{
		Path: "record/raw",
		Ext:  ".", // 默认h264扩展名为.h264,h265扩展名为.h265
	},
}

var plugin = InstallPlugin(RecordPluginConfig)

func (conf *RecordConfig) OnEvent(event any) {
	switch v := event.(type) {
	case FirstConfig, config.Config:
		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Hls.Init()
		conf.Raw.Init()
	case SEclose:
		streamPath := v.Stream.Path
		delete(conf.Flv.recording, streamPath)
		delete(conf.Mp4.recording, streamPath)
		delete(conf.Hls.recording, streamPath)
		delete(conf.Raw.recording, streamPath)
	case SEpublish:
		streamPath := v.Stream.Path
		if conf.Flv.NeedRecord(streamPath) {
			var flv FLVRecorder
			flv.IsInternal = true
			if flv.Start(streamPath) == nil {
				conf.Flv.recording[streamPath] = &flv
			}
		}
		if conf.Mp4.NeedRecord(streamPath) {
			if recoder := NewMP4Recorder(); recoder.Start(streamPath) == nil {
				conf.Mp4.recording[streamPath] = recoder
			}
		}
		if conf.Hls.NeedRecord(streamPath) {
			var hls HLSRecorder
			hls.IsInternal = true
			if hls.Start(streamPath) == nil {
				conf.Hls.recording[streamPath] = &hls
			}
		}
		if conf.Raw.NeedRecord(streamPath) {
			var raw RawRecorder
			raw.IsInternal = true
			if raw.Start(streamPath) == nil {
				conf.Raw.recording[streamPath] = &raw
			}
		}
	}
}
func (conf *RecordConfig) getRecorderConfigByType(t string) (recorder *Record) {
	switch t {
	case "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	}
	return
}

func getFLVDuration(file io.ReadSeeker) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
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
