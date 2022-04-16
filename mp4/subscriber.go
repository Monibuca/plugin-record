package mp4

import (
	"github.com/edgeware/mp4ff/aac"
	"github.com/edgeware/mp4ff/mp4"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/track"
	"m7s.live/engine/v4/util"
)

var defaultFtyp = mp4.NewFtyp("isom", 0x200, []string{
	"isom", "iso2", "avc1", "mp41",
})

type mediaContext struct {
	trackId  uint32
	fragment *mp4.Fragment
	ts       uint32 // 起始时间戳
}

func (m *mediaContext) push(recoder *Recorder, dt uint32, dur uint32, data []byte, flags uint32) {
	if m.fragment != nil && dt-m.ts > 5000 {
		m.fragment.Encode(recoder.fp)
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

type Recorder struct {
	Subscriber
	fp               config.FileWr
	*mp4.InitSegment `json:"-"`
	video            mediaContext
	audio            mediaContext
	seqNumber        uint32
}

func NewRecorder(fp config.FileWr) *Recorder {
	r := &Recorder{
		fp:          fp,
		InitSegment: mp4.CreateEmptyInit(),
	}
	r.InitSegment.Ftyp = defaultFtyp
	return r
}

func (r *Recorder) Close() error {
	if r.fp != nil {
		if r.video.fragment != nil {
			r.video.fragment.Encode(r.fp)
		}
		if r.audio.fragment != nil {
			r.audio.fragment.Encode(r.fp)
		}
		r.fp.Close()
	}
	return nil
}

func (r *Recorder) OnEvent(event any) {
	switch v := event.(type) {
	case *track.Video:
		moov := r.Moov
		trackID := uint32(len(moov.Traks) + 1)
		moov.Mvhd.NextTrackID = trackID + 1
		newTrak := mp4.CreateEmptyTrak(trackID, 1000, "video", "chi")
		moov.AddChild(newTrak)
		moov.Mvex.AddChild(mp4.CreateTrex(trackID))
		r.video.trackId = trackID
		switch v.CodecID {
		case codec.CodecID_H264:
			newTrak.SetAVCDescriptor("avc1", [][]byte{v.DecoderConfiguration.Raw[0]}, [][]byte{v.DecoderConfiguration.Raw[1]})
		case codec.CodecID_H265:
			newTrak.SetHEVCDescriptor("hev1", [][]byte{v.DecoderConfiguration.Raw[0]}, [][]byte{v.DecoderConfiguration.Raw[1]}, [][]byte{v.DecoderConfiguration.Raw[2]})
		}
		r.AddTrack(v)
	case *track.Audio:
		moov := r.Moov
		trackID := uint32(len(moov.Traks) + 1)
		moov.Mvhd.NextTrackID = trackID + 1
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
	case *AudioFrame:
		if r.audio.trackId != 0 {
			if r.InitSegment != nil {
				r.InitSegment.Ftyp.Encode(r.fp)
				r.InitSegment.Moov.Encode(r.fp)
				r.InitSegment = nil
			}
			r.audio.push(r, v.AbsTime-r.Audio.First.AbsTime, v.DeltaTime, util.ConcatBuffers(v.Raw), mp4.SyncSampleFlags)
		}
	case *VideoFrame:
		if r.video.trackId != 0 {
			flag := mp4.NonSyncSampleFlags
			if v.IFrame {
				flag = mp4.SyncSampleFlags
			}
			if r.InitSegment != nil {
				r.InitSegment.Ftyp.Encode(r.fp)
				r.InitSegment.Moov.Encode(r.fp)
				r.InitSegment = nil
			}
			r.video.push(r, v.AbsTime-r.Video.First.AbsTime, v.DeltaTime, util.ConcatBuffers(v.AVCC)[5:], flag)
		}

	default:
		r.Subscriber.OnEvent(event)
	}

}
