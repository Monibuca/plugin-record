package record

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgeware/mp4ff/aac"
	"github.com/edgeware/mp4ff/mp4"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
	"m7s.live/engine/v4/util"
)

var defaultFtyp = mp4.NewFtyp("isom", 0x200, []string{
	"isom", "iso2", "avc1", "mp41",
})

type mediaContext struct {
	trackId  uint32
	fragment *mp4.Fragment
	ts       uint32 // 每个小片段起始时间戳
	abs      uint32 // 绝对起始时间戳
	absSet   bool   // 是否设置过abs
}

func (m *mediaContext) push(recoder *MP4Recorder, dt uint32, dur uint32, data []byte, flags uint32) {
	if !m.absSet {
		m.abs = dt
		m.absSet = true
	}
	dt -= m.abs
	if m.fragment != nil && dt-m.ts > 5000 {
		m.fragment.Encode(recoder)
		m.fragment = nil
	}
	if m.fragment == nil {
		recoder.seqNumber++
		m.fragment, _ = mp4.CreateFragment(recoder.seqNumber, m.trackId)
		m.ts = dt
	}
	m.fragment.AddFullSample(mp4.FullSample{
		Data:       data,
		DecodeTime: uint64(dt),
		Sample: mp4.Sample{
			Flags: flags,
			Dur:   dur,
			Size:  uint32(len(data)),
		},
	})
}

type MP4Recorder struct {
	Recorder
	*mp4.InitSegment `json:"-"`
	video            mediaContext
	audio            mediaContext
	seqNumber        uint32
}

func NewMP4Recorder() *MP4Recorder {
	r := &MP4Recorder{
		InitSegment: mp4.CreateEmptyInit(),
	}
	r.Moov.Mvhd.NextTrackID = 1
	return r
}

func (r *MP4Recorder) Close() error {
	if r.Writer != nil {
		if r.video.fragment != nil {
			r.video.fragment.Encode(r.Writer)
			r.video.fragment = nil
		}
		if r.audio.fragment != nil {
			r.audio.fragment.Encode(r.Writer)
			r.audio.fragment = nil
		}
		r.Closer.Close()
	}
	return nil
}

func (r *MP4Recorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case *track.Video:
		moov := r.Moov
		trackID := moov.Mvhd.NextTrackID
		moov.Mvhd.NextTrackID++
		newTrak := mp4.CreateEmptyTrak(trackID, 1000, "video", "chi")
		moov.AddChild(newTrak)
		moov.Mvex.AddChild(mp4.CreateTrex(trackID))
		r.video.trackId = trackID
		switch v.CodecID {
		case codec.CodecID_H264:
			newTrak.SetAVCDescriptor("avc1", v.DecoderConfiguration.Raw[0:1], v.DecoderConfiguration.Raw[1:2], true)
		case codec.CodecID_H265:
			newTrak.SetHEVCDescriptor("hev1", v.DecoderConfiguration.Raw[0:1], v.DecoderConfiguration.Raw[1:2], v.DecoderConfiguration.Raw[2:3], true)
		}
		r.AddTrack(v)
	case *track.Audio:
		moov := r.Moov
		trackID := moov.Mvhd.NextTrackID
		moov.Mvhd.NextTrackID++
		newTrak := mp4.CreateEmptyTrak(trackID, 1000, "audio", "chi")
		moov.AddChild(newTrak)
		moov.Mvex.AddChild(mp4.CreateTrex(trackID))
		r.audio.trackId = trackID
		switch v.CodecID {
		case codec.CodecID_AAC:
			switch v.Profile {
			case 0:
				newTrak.SetAACDescriptor(aac.HEAACv1, int(v.SampleRate))
			case 1:
				newTrak.SetAACDescriptor(aac.AAClc, int(v.SampleRate))
			case 2:
				newTrak.SetAACDescriptor(aac.HEAACv2, int(v.SampleRate))
			}
		case codec.CodecID_PCMA:
			stsd := newTrak.Mdia.Minf.Stbl.Stsd
			pcma := mp4.CreateAudioSampleEntryBox("pcma",
				uint16(v.Channels),
				uint16(v.SampleSize), uint16(v.SampleRate), nil)
			stsd.AddChild(pcma)
		case codec.CodecID_PCMU:
			stsd := newTrak.Mdia.Minf.Stbl.Stsd
			pcmu := mp4.CreateAudioSampleEntryBox("pcmu",
				uint16(v.Channels),
				uint16(v.SampleSize), uint16(v.SampleRate), nil)
			stsd.AddChild(pcmu)
		}
		r.AddTrack(v)
	case ISubscriber:
		defaultFtyp.Encode(r)
		r.Moov.Encode(r)
		go r.Start()
	case *AudioFrame:
		if r.audio.trackId != 0 {
			r.audio.push(r, v.AbsTime, v.DeltaTime, util.ConcatBuffers(v.Raw), mp4.SyncSampleFlags)
		}
	case *VideoFrame:
		if r.Fragment != 0 && r.newFile {
			r.newFile = false
			r.Close()
			if file, err := r.CreateFileFn(filepath.Join(r.Stream.Path, strconv.FormatInt(time.Now().Unix(), 10)+r.Ext), false); err == nil {
				r.SetIO(file)
				r.audio.absSet = false
				r.video.absSet = false
				r.InitSegment = mp4.CreateEmptyInit()
				r.Moov.Mvhd.NextTrackID = 1
				r.OnEvent(r.Video.Track)
				r.OnEvent(r.Audio.Track)
				defaultFtyp.Encode(r)
				r.Moov.Encode(r)
				r.seqNumber = 0
			}
		}
		if r.video.trackId != 0 {
			flag := mp4.NonSyncSampleFlags
			if v.IFrame {
				flag = mp4.SyncSampleFlags
			}
			r.video.push(r, v.AbsTime - r.SkipTS, v.DeltaTime, util.ConcatBuffers(v.AVCC)[5:], flag)
		}
	}
}
