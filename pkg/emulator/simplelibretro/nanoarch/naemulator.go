package nanoarch

import (
	"github.com/disintegration/imaging"
	"image"
	"log"
	"sync"
	stdimage "image"
	"time"

	"github.com/giongto35/cloud-game/pkg/config"
	"github.com/giongto35/cloud-game/pkg/util"
)

/*
#include "libretro.h"
#cgo LDFLAGS: -ldl
#include <stdlib.h>
#include <stdio.h>
#include <dlfcn.h>
#include <string.h>

void bridge_retro_deinit(void *f);
unsigned bridge_retro_api_version(void *f);
void bridge_retro_get_system_info(void *f, struct retro_system_info *si);
void bridge_retro_get_system_av_info(void *f, struct retro_system_av_info *si);
bool bridge_retro_set_environment(void *f, void *callback);
void bridge_retro_set_video_refresh(void *f, void *callback);
void bridge_retro_set_input_poll(void *f, void *callback);
void bridge_retro_set_input_state(void *f, void *callback);
void bridge_retro_set_audio_sample(void *f, void *callback);
void bridge_retro_set_audio_sample_batch(void *f, void *callback);
bool bridge_retro_load_game(void *f, struct retro_game_info *gi);
void bridge_retro_run(void *f);
size_t bridge_retro_get_memory_size(void *f, unsigned id);
void* bridge_retro_get_memory_data(void *f, unsigned id);
bool bridge_retro_serialize(void *f, void *data, size_t size);
bool bridge_retro_unserialize(void *f, void *data, size_t size);
size_t bridge_retro_serialize_size(void *f);

bool coreEnvironment_cgo(unsigned cmd, void *data);
void coreVideoRefresh_cgo(void *data, unsigned width, unsigned height, size_t pitch);
void coreInputPoll_cgo();
void coreAudioSample_cgo(int16_t left, int16_t right);
size_t coreAudioSampleBatch_cgo(const int16_t *data, size_t frames);
int16_t coreInputState_cgo(unsigned port, unsigned device, unsigned index, unsigned id);
void coreLog_cgo(enum retro_log_level level, const char *msg);
*/
import "C"

const numAxes = 4

type constrollerState struct {
	keyState uint16
	axes     [numAxes]int16
}

// naEmulator implements CloudEmulator
type naEmulator struct {
	imageChannel chan<- *image.RGBA
	audioChannel chan<- []int16
	inputChannel <-chan InputEvent

	meta            config.EmulatorMeta
	gamePath        string
	roomID          string
	gameName        string
	isSavingLoading bool

	controllersMap map[string][]constrollerState
	done           chan struct{}

	// lock to lock uninteruptable operation
	lock *sync.Mutex
}

type InputEvent struct {
	RawState  []byte
	PlayerIdx int
	ConnID    string
}

var NAEmulator *naEmulator
var outputImg *image.RGBA

const maxPort = 8


// Init initialize new RetroArch cloud emulator
func Init(etype string, roomID string, inputChannel <-chan InputEvent) (*naEmulator, chan *image.RGBA, chan []int16) {
	imageChannel := make(chan *image.RGBA, 30)
	audioChannel := make(chan []int16, 30)
	meta := config.EmulatorMeta{
		Path:            "",
		Config:          "",
		Width:           240,
		Height:          160,
		AudioSampleRate: 0,
		Fps:             0,
		BaseWidth:       0,
		BaseHeight:      0,
		Ratio:           0,
		IsGlAllowed:     false,
		UsesLibCo:       false,
	}
	emulator := &naEmulator{
		meta:           meta,
		imageChannel:   imageChannel,
		audioChannel:   audioChannel,
		inputChannel:   inputChannel,
		controllersMap: map[string][]constrollerState{},
		roomID:         roomID,
		done:           make(chan struct{}, 1),
		lock:           &sync.Mutex{},
	}

	return emulator, imageChannel, audioChannel
}


func (na *naEmulator) LoadMeta(path string) config.EmulatorMeta {
	return na.meta
}

func (na *naEmulator) SetViewport(width int, height int) {
	// outputImg is tmp img used for decoding and reuse in encoding flow
	outputImg = image.NewRGBA(image.Rect(0, 0, width, height))
}

func (na *naEmulator) Start() {
	na.playGame(na.gamePath)
	ticker := time.NewTicker(time.Second / 60)

	for range ticker.C {
		select {
		// Slow response here
		case <-na.done:
			close(na.imageChannel)
			close(na.audioChannel)
			log.Println("Closed Director")
			return
		default:
			im := stdimage.NewRGBA(stdimage.Rect(0, 0, int(na.meta.Width), int(na.meta.Height)))
			na.imageChannel <- im
			p := make([]int16, 10)
			na.audioChannel <- p
		}

		na.GetLock()
		na.ReleaseLock()
	}
}

func (na *naEmulator) playGame(path string) {
	// When start game, we also try loading if there was a saved state
	na.LoadGame()
}

func (na *naEmulator) SaveGame(saveExtraFunc func() error) error {
	if na.roomID != "" {
		err := na.Save()
		if err != nil {
			return err
		}
		err = saveExtraFunc()
		if err != nil {
			return err
		}
	}

	return nil
}

func (na *naEmulator) LoadGame() error {
	if na.roomID != "" {
		err := na.Load()
		if err != nil {
			log.Println("Error: Cannot load", err)
			return err
		}
	}

	return nil
}

func (na *naEmulator) GetHashPath() string {
	return util.GetSavePath(na.roomID)
}

func (na *naEmulator) Close() {
	// Unload and deinit in the core.
	close(na.done)
}
