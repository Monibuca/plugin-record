package record

import (
	"net"

	"github.com/yapingcat/gomedia/go-mp4"
	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/util"
)

type MP4Recorder struct {
	Recorder
	*mp4.Movmuxer `json:"-" yaml:"-"`
	videoId       uint32
	audioId       uint32
}

func NewMP4Recorder() *MP4Recorder {
	r := &MP4Recorder{}
	r.Record = RecordPluginConfig.Mp4
	return r
}

func (r *MP4Recorder) Start(streamPath string) (err error) {
	r.ID = streamPath + "/mp4"
	return r.start(r, streamPath, SUBTYPE_RAW)
}

func (r *MP4Recorder) Close() (err error) {
	if r.File != nil {
		err = r.Movmuxer.WriteTrailer()
		if err != nil {
			r.Error("mp4 write trailer", zap.Error(err))
		} else {
			// _, err = r.file.Write(r.cache.buf)
			r.Info("mp4 write trailer", zap.Error(err))
		}
		err = r.File.Close()
	}
	return
}
func (r *MP4Recorder) setTracks() {
	if r.Audio != nil {
		switch r.Audio.CodecID {
		case codec.CodecID_AAC:
			r.audioId = r.AddAudioTrack(mp4.MP4_CODEC_AAC, mp4.WithExtraData(r.Audio.SequenceHead[2:]))
		case codec.CodecID_PCMA:
			r.audioId = r.AddAudioTrack(mp4.MP4_CODEC_G711A)
		case codec.CodecID_PCMU:
			r.audioId = r.AddAudioTrack(mp4.MP4_CODEC_G711U)
		}
	}
	if r.Video != nil {
		switch r.Video.CodecID {
		case codec.CodecID_H264:
			r.videoId = r.AddVideoTrack(mp4.MP4_CODEC_H264, mp4.WithExtraData(r.Video.SequenceHead[5:]))
		case codec.CodecID_H265:
			r.videoId = r.AddVideoTrack(mp4.MP4_CODEC_H265, mp4.WithExtraData(r.Video.SequenceHead[5:]))
		}
	}
}
func (r *MP4Recorder) OnEvent(event any) {
	var err error
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case FileWr:
		r.Movmuxer, err = mp4.CreateMp4Muxer(v)
		if err != nil {
			r.Error("mp4 create muxer", zap.Error(err))
		} else {
			r.setTracks()
		}
	case AudioFrame:
		if r.audioId != 0 {
			var audioData []byte
			if v.ADTS == nil {
				audioData = v.AUList.ToBytes()
			} else {
				audioData = util.ConcatBuffers(append(net.Buffers{v.ADTS.Value}, v.AUList.ToBuffers()...))
			}
			r.Write(r.audioId, audioData, uint64(v.AbsTime+(v.PTS-v.DTS)/90), uint64(v.AbsTime))
		}
	case VideoFrame:
		if r.videoId != 0 {
			r.Write(r.videoId, util.ConcatBuffers(v.GetAnnexB()), uint64(v.AbsTime+(v.PTS-v.DTS)/90), uint64(v.AbsTime))
		}
	}
}
