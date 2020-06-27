package nanoarch

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"image"
	"log"
	"sync"
	"time"

	"github.com/giongto35/cloud-game/pkg/config"
	"github.com/giongto35/cloud-game/pkg/util"
	"github.com/giongto35/cloud-game/pkg/util/gamelist/2048"

)



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
		BaseWidth:       240,
		BaseHeight:      160,
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
	game,err := twenty48.NewGame()
	go game.Run()
	img, _ := ebiten.NewImage(na.meta.Width, na.meta.Height, ebiten.FilterDefault)
	if err != nil{
		log.Fatal(err)
	}
	ticker := time.NewTicker(time.Second / 60)

	for {
		fmt.Println("In start, ticker...")
		<- ticker.C
		select {
		// Slow response here
		case <-na.done:
			close(na.imageChannel)
			close(na.audioChannel)
			log.Println("Closed Director")
			return
		default:
			game.Draw(img)
			na.imageChannel <- img.ToRGBA()
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
