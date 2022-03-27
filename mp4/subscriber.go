package mp4

import (
	"bytes"
	"io"
	"net"

	"github.com/yapingcat/gomedia/mp4"
	"github.com/yapingcat/gomedia/mpeg"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/track"
)

type Recorder struct {
	Subscriber
	muxer *mp4.Movmuxer
	vid   uint32
	aid   uint32
	fp    config.FileWr
}

func NewRecorder(fp config.FileWr) *Recorder {
	r := &Recorder{
		fp: fp,
	}
	r.muxer = mp4.CreateMp4Muxer(r)
	return r
}
func (r *Recorder) Write(p []byte) (n int, err error) {
	return r.fp.Write(p)
}
func (r *Recorder) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}
func (r *Recorder) Tell() (offset int64) {
	offset, _ = r.fp.Seek(0, io.SeekCurrent)
	return
}
func (r *Recorder) Close() error {
	if r.fp != nil {
		r.fp.Close()
		return r.muxer.Writetrailer()
	}
	return nil
}

func (r *Recorder) OnEvent(event any) {
	switch v := event.(type) {
	case *track.Video:
		switch v.CodecID {
		case codec.CodecID_H264:
			r.vid = r.muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
		case codec.CodecID_H265:
			r.vid = r.muxer.AddVideoTrack(mp4.MP4_CODEC_H265)
		}
		r.AddTrack(v)
	case *track.Audio:
		switch v.CodecID {
		case codec.CodecID_AAC:
			r.aid = r.muxer.AddAudioTrack(mp4.MP4_CODEC_AAC, v.Channels, v.SampleSize, uint(v.SampleRate))
		case codec.CodecID_PCMA:
			r.aid = r.muxer.AddAudioTrack(mp4.MP4_CODEC_G711A, v.Channels, v.SampleSize, uint(v.SampleRate))
		case codec.CodecID_PCMU:
			r.aid = r.muxer.AddAudioTrack(mp4.MP4_CODEC_G711U, v.Channels, v.SampleSize, uint(v.SampleRate))
		}
		r.AddTrack(v)
	case *AudioFrame:
		if r.aid != 0 {
			var buf bytes.Buffer
			if r.AudioTrack.CodecID == codec.CodecID_AAC {
				buf.Grow(7)
				adts := buf.Next(7)
				for _, s := range v.Raw {
					buf.Write(s)
				}
				copy(adts, mpeg.ConvertASCToADTS(r.AudioTrack.DecoderConfiguration.Raw, buf.Len()))
			} else {
				for _, s := range v.Raw {
					buf.Write(s)
				}
			}
			r.muxer.Write(r.aid, buf.Bytes(), uint64(v.PTS), uint64(v.DTS))
		}
	case *VideoFrame:
		if r.vid != 0 {
			var buf bytes.Buffer
			if v.IFrame {
				for _, nalu := range r.VideoTrack.DecoderConfiguration.Raw {
					buf.Write(codec.NALU_Delimiter2)
					buf.Write(nalu)
				}
			}
			buf.Write(codec.NALU_Delimiter2)
			var b net.Buffers = net.Buffers(v.Raw[0])
			b.WriteTo(&buf)
			for _, nalu := range v.Raw[1:] {
				buf.Write(codec.NALU_Delimiter1)
				b = net.Buffers(nalu)
				b.WriteTo(&buf)
			}
			r.muxer.Write(r.vid, buf.Bytes(), uint64(v.PTS), uint64(v.DTS))
		}
	default:
		r.Subscriber.OnEvent(event)
	}

}
