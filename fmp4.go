package record

import (
	"github.com/edgeware/mp4ff/aac"
	"github.com/edgeware/mp4ff/mp4"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
)

type mediaContext struct {
	trackId  uint32
	fragment *mp4.Fragment
	ts       uint32 // 每个小片段起始时间戳
}

func (m *mediaContext) push(recoder *FMP4Recorder, dt uint32, dur uint32, data []byte, flags uint32) {
	if m.fragment != nil && dt-m.ts > 1000 {
		m.fragment.Encode(recoder.File)
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

type FMP4Recorder struct {
	Recorder
	initSegment *mp4.InitSegment `json:"-" yaml:"-"`
	video       mediaContext
	audio       mediaContext
	seqNumber   uint32
	ftyp        *mp4.FtypBox
}

func NewFMP4Recorder() *FMP4Recorder {
	r := &FMP4Recorder{}
	r.Record = RecordPluginConfig.Fmp4
	return r
}

func (r *FMP4Recorder) Start(streamPath string) (err error) {
	r.ID = streamPath + "/fmp4"
	return r.start(r, streamPath, SUBTYPE_RAW)
}

func (r *FMP4Recorder) Close() error {
	if r.File != nil {
		if r.video.fragment != nil {
			r.video.fragment.Encode(r.File)
			r.video.fragment = nil
		}
		if r.audio.fragment != nil {
			r.audio.fragment.Encode(r.File)
			r.audio.fragment = nil
		}
		r.File.Close()
	}
	return nil
}

func (r *FMP4Recorder) OnEvent(event any) {
	r.Recorder.OnEvent(event)
	switch v := event.(type) {
	case FileWr:
		r.initSegment = mp4.CreateEmptyInit()
		r.initSegment.Moov.Mvhd.NextTrackID = 1
		if r.VideoReader != nil {
			moov := r.initSegment.Moov
			trackID := moov.Mvhd.NextTrackID
			moov.Mvhd.NextTrackID++
			newTrak := mp4.CreateEmptyTrak(trackID, 1000, "video", "chi")
			moov.AddChild(newTrak)
			moov.Mvex.AddChild(mp4.CreateTrex(trackID))
			r.video.trackId = trackID
			switch r.Video.CodecID {
			case codec.CodecID_H264:
				r.ftyp = mp4.NewFtyp("isom", 0x200, []string{
					"isom", "iso2", "avc1", "mp41",
				})
				newTrak.SetAVCDescriptor("avc1", r.Video.ParamaterSets[0:1], r.Video.ParamaterSets[1:2], true)
			case codec.CodecID_H265:
				r.ftyp = mp4.NewFtyp("isom", 0x200, []string{
					"isom", "iso2", "hvc1", "mp41",
				})
				newTrak.SetHEVCDescriptor("hvc1", r.Video.ParamaterSets[0:1], r.Video.ParamaterSets[1:2], r.Video.ParamaterSets[2:3], true)
			}
		}
		if r.AudioReader != nil {
			moov := r.initSegment.Moov
			trackID := moov.Mvhd.NextTrackID
			moov.Mvhd.NextTrackID++
			newTrak := mp4.CreateEmptyTrak(trackID, 1000, "audio", "chi")
			moov.AddChild(newTrak)
			moov.Mvex.AddChild(mp4.CreateTrex(trackID))
			r.audio.trackId = trackID
			switch r.Audio.CodecID {
			case codec.CodecID_AAC:
				switch r.Audio.AudioObjectType {
				case 1:
					newTrak.SetAACDescriptor(aac.HEAACv1, int(r.Audio.SampleRate))
				case 2:
					newTrak.SetAACDescriptor(aac.AAClc, int(r.Audio.SampleRate))
				case 3:
					newTrak.SetAACDescriptor(aac.HEAACv2, int(r.Audio.SampleRate))
				}
			case codec.CodecID_PCMA:
				stsd := newTrak.Mdia.Minf.Stbl.Stsd
				pcma := mp4.CreateAudioSampleEntryBox("pcma",
					uint16(r.Audio.Channels),
					uint16(r.Audio.SampleSize), uint16(r.Audio.SampleRate), nil)
				stsd.AddChild(pcma)
			case codec.CodecID_PCMU:
				stsd := newTrak.Mdia.Minf.Stbl.Stsd
				pcmu := mp4.CreateAudioSampleEntryBox("pcmu",
					uint16(r.Audio.Channels),
					uint16(r.Audio.SampleSize), uint16(r.Audio.SampleRate), nil)
				stsd.AddChild(pcmu)
			}
		}
		r.ftyp.Encode(v)
		r.initSegment.Moov.Encode(v)
		r.seqNumber = 0
	case AudioFrame:
		if r.audio.trackId != 0 {
			r.audio.push(r, v.AbsTime, v.DeltaTime, v.AUList.ToBytes(), mp4.SyncSampleFlags)
		}
	case VideoFrame:
		if r.video.trackId != 0 {
			flag := mp4.NonSyncSampleFlags
			if v.IFrame {
				flag = mp4.SyncSampleFlags
			}
			if data := v.AVCC.ToBytes(); len(data) > 5 {
				r.video.push(r, v.AbsTime, v.DeltaTime, data[5:], flag)
			}
		}
	}
}
