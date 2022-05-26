package record

import (
	"io"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/codec/mpegts"
	"m7s.live/plugin/hls/v4"
)

type HLSRecorder struct {
	playlist           hls.Playlist
	asc                codec.AudioSpecificConfig
	video_cc, audio_cc uint16
	packet             mpegts.MpegTsPESPacket
	Recorder
	tsWriter io.WriteCloser
}

func (h *HLSRecorder) OnEvent(event any) {
	var err error
	defer func() {
		if err != nil {
			h.Error("HLSRecorder Stop", zap.Error(err))
			h.Stop()
		}
	}()
	h.Recorder.OnEvent(event)
	switch v := event.(type) {
	case *HLSRecorder:
		h.playlist = hls.Playlist{
			Writer:         h.Writer,
			Version:        3,
			Sequence:       0,
			Targetduration: h.Fragment * 1000,
		}
		if err = h.playlist.Init(); err != nil {
			return
		}
		if err = h.createHlsTsSegmentFile(); err != nil {
			return
		}
	case AudioDeConf:
		h.asc, err = hls.DecodeAudioSpecificConfig(v.AVCC[0])
	case *AudioFrame:
		if h.packet, err = hls.AudioPacketToPES(v, h.asc); err != nil {
			return
		}
		pes := &mpegts.MpegtsPESFrame{
			Pid:                       0x102,
			IsKeyFrame:                false,
			ContinuityCounter:         byte(h.audio_cc % 16),
			ProgramClockReferenceBase: uint64(v.DTS),
		}
		//frame.ProgramClockReferenceBase = 0
		if err = mpegts.WritePESPacket(h.tsWriter, pes, h.packet); err != nil {
			return
		}
		h.audio_cc = uint16(pes.ContinuityCounter)
	case *VideoFrame:
		h.packet, err = hls.VideoPacketToPES(v, h.Video.Track.DecoderConfiguration)
		if err != nil {
			return
		}
		if h.Fragment != 0 {
			if h.newFile {
				h.newFile = false
				h.createHlsTsSegmentFile()
			}
		}
		pes := &mpegts.MpegtsPESFrame{
			Pid:                       0x101,
			IsKeyFrame:                v.IFrame,
			ContinuityCounter:         byte(h.video_cc % 16),
			ProgramClockReferenceBase: uint64(v.DTS),
		}
		if err = mpegts.WritePESPacket(h.tsWriter, pes, h.packet); err != nil {
			return
		}
		h.video_cc = uint16(pes.ContinuityCounter)
	}

}

// 创建一个新的ts文件
func (h *HLSRecorder) createHlsTsSegmentFile() (err error) {
	tsFilename := strconv.FormatInt(time.Now().Unix(), 10) + ".ts"
	fw, err := h.CreateFileFn(filepath.Join(h.Stream.Path, tsFilename), false)
	if err != nil {
		return err
	}
	h.tsWriter = fw
	inf := hls.PlaylistInf{
		Duration: float64(h.Fragment),
		Title:    tsFilename,
	}
	if err = h.playlist.WriteInf(inf); err != nil {
		return
	}
	if err = mpegts.WriteDefaultPATPacket(fw); err != nil {
		return err
	}
	if err = mpegts.WriteDefaultPMTPacket(fw); err != nil {
		return err
	}
	return err
}
