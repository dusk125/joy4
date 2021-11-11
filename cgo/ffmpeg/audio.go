package ffmpeg

/*
#include "ffmpeg.h"
int wrap_avcodec_decode_audio4(AVCodecContext *ctx, AVFrame *frame, void *data, int size, int *got) {
	struct AVPacket pkt = {.data = data, .size = size};
	return avcodec_decode_audio4(ctx, frame, got, &pkt);
}
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/dusk125/joy4/av"
	"github.com/dusk125/joy4/codec/aacparser"
)

const debug = false

func sampleFormatAV2FF(sampleFormat av.SampleFormat) (ffsamplefmt int32) {
	switch sampleFormat {
	case av.U8:
		ffsamplefmt = C.AV_SAMPLE_FMT_U8
	case av.S16:
		ffsamplefmt = C.AV_SAMPLE_FMT_S16
	case av.S32:
		ffsamplefmt = C.AV_SAMPLE_FMT_S32
	case av.FLT:
		ffsamplefmt = C.AV_SAMPLE_FMT_FLT
	case av.DBL:
		ffsamplefmt = C.AV_SAMPLE_FMT_DBL
	case av.U8P:
		ffsamplefmt = C.AV_SAMPLE_FMT_U8P
	case av.S16P:
		ffsamplefmt = C.AV_SAMPLE_FMT_S16P
	case av.S32P:
		ffsamplefmt = C.AV_SAMPLE_FMT_S32P
	case av.FLTP:
		ffsamplefmt = C.AV_SAMPLE_FMT_FLTP
	case av.DBLP:
		ffsamplefmt = C.AV_SAMPLE_FMT_DBLP
	}
	return
}

func sampleFormatFF2AV(ffsamplefmt int32) (sampleFormat av.SampleFormat) {
	switch ffsamplefmt {
	case C.AV_SAMPLE_FMT_U8: ///< unsigned 8 bits
		sampleFormat = av.U8
	case C.AV_SAMPLE_FMT_S16: ///< signed 16 bits
		sampleFormat = av.S16
	case C.AV_SAMPLE_FMT_S32: ///< signed 32 bits
		sampleFormat = av.S32
	case C.AV_SAMPLE_FMT_FLT: ///< float
		sampleFormat = av.FLT
	case C.AV_SAMPLE_FMT_DBL: ///< double
		sampleFormat = av.DBL
	case C.AV_SAMPLE_FMT_U8P: ///< unsigned 8 bits, planar
		sampleFormat = av.U8P
	case C.AV_SAMPLE_FMT_S16P: ///< signed 16 bits, planar
		sampleFormat = av.S16P
	case C.AV_SAMPLE_FMT_S32P: ///< signed 32 bits, planar
		sampleFormat = av.S32P
	case C.AV_SAMPLE_FMT_FLTP: ///< float, planar
		sampleFormat = av.FLTP
	case C.AV_SAMPLE_FMT_DBLP: ///< double, planar
		sampleFormat = av.DBLP
	}
	return
}

func audioFrameAssignToAVParams(f *C.AVFrame, frame *av.AudioFrame) {
	frame.SampleFormat = sampleFormatFF2AV(int32(f.format))
	frame.ChannelLayout = channelLayoutFF2AV(f.channel_layout)
	frame.SampleRate = int(f.sample_rate)
}

func audioFrameAssignToAVData(f *C.AVFrame, frame *av.AudioFrame) {
	frame.SampleCount = int(f.nb_samples)
	frame.Data = make([][]byte, int(f.channels))
	for i := 0; i < int(f.channels); i++ {
		frame.Data[i] = C.GoBytes(unsafe.Pointer(f.data[i]), f.linesize[0])
	}
}

func audioFrameAssignToAV(f *C.AVFrame, frame *av.AudioFrame) {
	audioFrameAssignToAVParams(f, frame)
	audioFrameAssignToAVData(f, frame)
}

func channelLayoutFF2AV(layout C.uint64_t) (channelLayout av.ChannelLayout) {
	if layout&C.AV_CH_FRONT_CENTER != 0 {
		channelLayout |= av.CH_FRONT_CENTER
	}
	if layout&C.AV_CH_FRONT_LEFT != 0 {
		channelLayout |= av.CH_FRONT_LEFT
	}
	if layout&C.AV_CH_FRONT_RIGHT != 0 {
		channelLayout |= av.CH_FRONT_RIGHT
	}
	if layout&C.AV_CH_BACK_CENTER != 0 {
		channelLayout |= av.CH_BACK_CENTER
	}
	if layout&C.AV_CH_BACK_LEFT != 0 {
		channelLayout |= av.CH_BACK_LEFT
	}
	if layout&C.AV_CH_BACK_RIGHT != 0 {
		channelLayout |= av.CH_BACK_RIGHT
	}
	if layout&C.AV_CH_SIDE_LEFT != 0 {
		channelLayout |= av.CH_SIDE_LEFT
	}
	if layout&C.AV_CH_SIDE_RIGHT != 0 {
		channelLayout |= av.CH_SIDE_RIGHT
	}
	if layout&C.AV_CH_LOW_FREQUENCY != 0 {
		channelLayout |= av.CH_LOW_FREQ
	}
	return
}

func channelLayoutAV2FF(channelLayout av.ChannelLayout) (layout C.uint64_t) {
	if channelLayout&av.CH_FRONT_CENTER != 0 {
		layout |= C.AV_CH_FRONT_CENTER
	}
	if channelLayout&av.CH_FRONT_LEFT != 0 {
		layout |= C.AV_CH_FRONT_LEFT
	}
	if channelLayout&av.CH_FRONT_RIGHT != 0 {
		layout |= C.AV_CH_FRONT_RIGHT
	}
	if channelLayout&av.CH_BACK_CENTER != 0 {
		layout |= C.AV_CH_BACK_CENTER
	}
	if channelLayout&av.CH_BACK_LEFT != 0 {
		layout |= C.AV_CH_BACK_LEFT
	}
	if channelLayout&av.CH_BACK_RIGHT != 0 {
		layout |= C.AV_CH_BACK_RIGHT
	}
	if channelLayout&av.CH_SIDE_LEFT != 0 {
		layout |= C.AV_CH_SIDE_LEFT
	}
	if channelLayout&av.CH_SIDE_RIGHT != 0 {
		layout |= C.AV_CH_SIDE_RIGHT
	}
	if channelLayout&av.CH_LOW_FREQ != 0 {
		layout |= C.AV_CH_LOW_FREQUENCY
	}
	return
}

type AudioDecoder struct {
	ff            *ffctx
	ChannelLayout av.ChannelLayout
	SampleFormat  av.SampleFormat
	SampleRate    int
	Extradata     []byte
}

func (self *AudioDecoder) Setup() (err error) {
	ff := &self.ff.ff

	ff.frame = C.av_frame_alloc()

	if len(self.Extradata) > 0 {
		ff.codecCtx.extradata = (*C.uint8_t)(unsafe.Pointer(&self.Extradata[0]))
		ff.codecCtx.extradata_size = C.int(len(self.Extradata))
	}
	if debug {
		fmt.Println("ffmpeg: Decoder.Setup Extradata.len", len(self.Extradata))
	}

	ff.codecCtx.sample_rate = C.int(self.SampleRate)
	ff.codecCtx.channel_layout = channelLayoutAV2FF(self.ChannelLayout)
	ff.codecCtx.channels = C.int(self.ChannelLayout.Count())
	if C.avcodec_open2(ff.codecCtx, ff.codec, nil) != 0 {
		err = fmt.Errorf("ffmpeg: decoder: avcodec_open2 failed")
		return
	}
	self.SampleFormat = sampleFormatFF2AV(ff.codecCtx.sample_fmt)
	self.ChannelLayout = channelLayoutFF2AV(ff.codecCtx.channel_layout)
	if self.SampleRate == 0 {
		self.SampleRate = int(ff.codecCtx.sample_rate)
	}

	return
}

func (self *AudioDecoder) Decode(pkt []byte) (gotframe bool, frame av.AudioFrame, err error) {
	ff := &self.ff.ff

	cgotframe := C.int(0)
	cerr := C.wrap_avcodec_decode_audio4(ff.codecCtx, ff.frame, unsafe.Pointer(&pkt[0]), C.int(len(pkt)), &cgotframe)
	if cerr < C.int(0) {
		err = fmt.Errorf("ffmpeg: avcodec_decode_audio4 failed: %d", cerr)
		return
	}

	if cgotframe != C.int(0) {
		gotframe = true
		audioFrameAssignToAV(ff.frame, &frame)
		frame.SampleRate = self.SampleRate

		if debug {
			fmt.Println("ffmpeg: Decode", frame.SampleCount, frame.SampleRate, frame.ChannelLayout, frame.SampleFormat)
		}
	}

	return
}

func (self *AudioDecoder) Close() {
	freeFFCtx(self.ff)
}

func NewAudioDecoder(codec av.AudioCodecData) (dec *AudioDecoder, err error) {
	_dec := &AudioDecoder{}
	var id uint32

	switch codec.Type() {
	case av.AAC:
		if aaccodec, ok := codec.(aacparser.CodecData); ok {
			_dec.Extradata = aaccodec.MPEG4AudioConfigBytes()
			id = C.AV_CODEC_ID_AAC
		} else {
			err = fmt.Errorf("ffmpeg: aac CodecData must be aacparser.CodecData")
			return
		}

	case av.SPEEX:
		id = C.AV_CODEC_ID_SPEEX

	case av.PCM_MULAW:
		id = C.AV_CODEC_ID_PCM_MULAW

	case av.PCM_ALAW:
		id = C.AV_CODEC_ID_PCM_ALAW

	default:
		if ffcodec, ok := codec.(audioCodecData); ok {
			_dec.Extradata = ffcodec.extradata
			id = ffcodec.codecId
		} else {
			err = fmt.Errorf("ffmpeg: invalid CodecData for ffmpeg to decode")
			return
		}
	}

	c := C.avcodec_find_decoder(id)
	if c == nil || C.avcodec_get_type(c.id) != C.AVMEDIA_TYPE_AUDIO {
		err = fmt.Errorf("ffmpeg: cannot find audio decoder id=%d", id)
		return
	}

	if _dec.ff, err = newFFCtxByCodec(c); err != nil {
		return
	}

	_dec.SampleFormat = codec.SampleFormat()
	_dec.SampleRate = codec.SampleRate()
	_dec.ChannelLayout = codec.ChannelLayout()
	if err = _dec.Setup(); err != nil {
		return
	}

	dec = _dec
	return
}

type audioCodecData struct {
	codecId       uint32
	sampleFormat  av.SampleFormat
	channelLayout av.ChannelLayout
	sampleRate    int
	extradata     []byte
}

func (self audioCodecData) Type() av.CodecType {
	return av.MakeAudioCodecType(self.codecId)
}

func (self audioCodecData) SampleRate() int {
	return self.sampleRate
}

func (self audioCodecData) SampleFormat() av.SampleFormat {
	return self.sampleFormat
}

func (self audioCodecData) ChannelLayout() av.ChannelLayout {
	return self.channelLayout
}

func (self audioCodecData) PacketDuration(data []byte) (dur time.Duration, err error) {
	// TODO: implement it: ffmpeg get_audio_frame_duration
	err = fmt.Errorf("ffmpeg: cannot get packet duration")
	return
}
