package h264encoder

import (
	"bytes"
	"image"
	"log"

	"github.com/gen2brain/x264-go"
)

//import (
//"fmt"
//"log"
//"time"
//"unsafe"

//"github.com/giongto35/cloud-game/config"
//)

//// https://chromium.googlesource.com/webm/libvpx/+/master/examples/simple_encoder.c

//[>
//#cgo pkg-config: vpx
//#include <stdlib.h>
//#include "vpx/vpx_encoder.h"
//#include "tools_common.h"

//typedef struct GoBytes {
//void *bs;
//int size;
//} GoBytesType;

//vpx_codec_err_t call_vpx_codec_enc_config_default(const VpxInterface *encoder, vpx_codec_enc_cfg_t *cfg) {
//return vpx_codec_enc_config_default(encoder->codec_interface(), cfg, 0);
//}
//vpx_codec_err_t call_vpx_codec_enc_init(vpx_codec_ctx_t *codec, const VpxInterface *encoder, vpx_codec_enc_cfg_t *cfg) {
//return vpx_codec_enc_init(codec, encoder->codec_interface(), cfg, 0);
//}
//GoBytesType get_frame_buffer(vpx_codec_ctx_t *codec, vpx_codec_iter_t *iter) {
//// iter has set to NULL when after add new image
//GoBytesType bytes = {NULL, 0};
//const vpx_codec_cx_pkt_t *pkt = vpx_codec_get_cx_data(codec, iter);
//if (pkt != NULL && pkt->kind == VPX_CODEC_CX_FRAME_PKT) {
//bytes.bs = pkt->data.frame.buf;
//bytes.size = pkt->data.frame.sz;
//}
//return bytes;
//}
//*/
//import "C"

const chanSize = 2

// NewVpxEncoder create h264 encoder
func NewVpxEncoder(width, height, fps int) (*VpxEncoder, error) {
	v := &VpxEncoder{
		Output: make(chan []byte, 5*chanSize),
		Input:  make(chan *image.RGBA, chanSize),

		IsRunning: true,
		Done:      false,

		buf:    bytes.NewBuffer(make([]byte, 0)),
		width:  width,
		height: height,
		fps:    fps,
	}

	if err := v.init(); err != nil {
		return nil, err
	}

	return v, nil
}

// VpxEncoder yuvI420 image to vp8 video
type VpxEncoder struct {
	Output chan []byte      // frame
	Input  chan *image.RGBA // yuvI420

	buf *bytes.Buffer
	enc *x264.Encoder

	IsRunning bool
	Done      bool
	// C
	width  int
	height int
	fps    int
}

func (v *VpxEncoder) init() error {
	v.IsRunning = true

	opts := &x264.Options{
		Width:     v.width,
		Height:    v.height,
		FrameRate: v.fps,
		Tune:      "zerolatency",
		Preset:    "veryfast",
		Profile:   "baseline",
		//LogLevel:  x264.LogDebug,
	}

	enc, err := x264.NewEncoder(v.buf, opts)
	if err != nil {
		panic(err)
	}
	v.enc = enc

	go v.startLooping()
	return nil
}

func (v *VpxEncoder) startLooping() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Warn: Recovered panic in encoding ", r)
		}
	}()

	for img := range v.Input {
		if v.Done == true {
			// The first time we see IsRunning set to false, we release and return
			v.Release()
			return
		}

		v.enc.Encode(img)
		v.Output <- v.buf.Bytes()
		v.buf.Reset()
	}

	if v.Done == true {
		// The first time we see IsRunning set to false, we release and return
		v.Release()
		return
	}
}

// Release release memory and stop loop
func (v *VpxEncoder) Release() {
	if v.IsRunning {
		v.IsRunning = false
		log.Println("Releasing encoder")
		// TODO: Bug here, after close it will signal
		close(v.Output)
		if v.Input != nil {
			close(v.Input)
		}
		err := v.enc.Close()
		if err != nil {
			log.Println("Failed to close H264 encoder")
		}
	}
	// TODO: Can we merge IsRunning and Done together
}
