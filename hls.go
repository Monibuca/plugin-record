package record

import (
	"math"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/codec/mpegts"
	"m7s.live/engine/v4/util"
	"m7s.live/plugin/hls/v4"
)

type HLSRecorder struct {
	playlist           hls.Playlist
	video_cc, audio_cc byte
	packet             mpegts.MpegTsPESPacket
	Recorder
	MemoryTs
}

func NewHLSRecorder() (r *HLSRecorder) {
	r = &HLSRecorder{}
	r.Record = RecordPluginConfig.Hls
	return r
}

func (h *HLSRecorder) Start(streamPath string) error {
	h.ID = streamPath + "/hls"
	return h.start(h, streamPath, SUBTYPE_RAW)
}

func (h *HLSRecorder) OnEvent(event any) {
	var err error
	defer func() {
		if err != nil {
			h.Stop(zap.Error(err))
		}
	}()
	switch v := event.(type) {
	case *HLSRecorder:
		h.BytesPool = make(util.BytesPool, 17)
		if h.Writer, err = h.createFile(); err != nil {
			return
		}
		h.SetIO(h.Writer)
		h.playlist = hls.Playlist{
			Writer:         h.Writer,
			Version:        3,
			Sequence:       0,
			Targetduration: int(math.Ceil(h.Fragment.Seconds())),
		}
		if err = h.playlist.Init(); err != nil {
			return
		}
		if h.File, err = h.CreateFile(); err != nil {
			return
		}
	case AudioFrame:
		h.Recorder.OnEvent(event)
		pes := &mpegts.MpegtsPESFrame{
			Pid:                       mpegts.PID_AUDIO,
			IsKeyFrame:                false,
			ContinuityCounter:         h.audio_cc,
			ProgramClockReferenceBase: uint64(v.DTS),
		}
		h.WriteAudioFrame(v, pes)
		h.BLL.WriteTo(h.File)
		h.Recycle()
		h.Clear()
		h.audio_cc = pes.ContinuityCounter
	case VideoFrame:
		h.Recorder.OnEvent(event)
		pes := &mpegts.MpegtsPESFrame{
			Pid:                       mpegts.PID_VIDEO,
			IsKeyFrame:                v.IFrame,
			ContinuityCounter:         h.video_cc,
			ProgramClockReferenceBase: uint64(v.DTS),
		}
		if err = h.WriteVideoFrame(v, pes); err != nil {
			return
		}
		h.BLL.WriteTo(h.File)
		h.Recycle()
		h.Clear()
		h.video_cc = pes.ContinuityCounter
	default:
		h.Recorder.OnEvent(v)
	}
}

// 创建一个新的ts文件
func (h *HLSRecorder) CreateFile() (fw FileWr, err error) {
	tsFilename := strconv.FormatInt(time.Now().Unix(), 10) + ".ts"
	filePath := filepath.Join(h.Stream.Path, tsFilename)
	fw, err = h.CreateFileFn(filePath, false)
	if err != nil {
		h.Error("create file", zap.String("path", filePath), zap.Error(err))
		return
	}
	h.Info("create file", zap.String("path", filePath))
	inf := hls.PlaylistInf{
		Duration: h.Fragment.Seconds(),
		Title:    tsFilename,
	}
	if err = h.playlist.WriteInf(inf); err != nil {
		return
	}
	if err = mpegts.WriteDefaultPATPacket(fw); err != nil {
		return
	}
	var vcodec codec.VideoCodecID = 0
	var acodec codec.AudioCodecID = 0
	if h.Video != nil {
		vcodec = h.Video.CodecID
	}
	if h.Audio != nil {
		acodec = h.Audio.CodecID
	}
	mpegts.WritePMTPacket(fw, vcodec, acodec)
	return
}
