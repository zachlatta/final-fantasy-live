package emulator

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/gordonklaus/portaudio"
	"github.com/paked/nes/ui"
)

var DefaultSettings = Settings{
	Width:  246,
	Height: 240,
	Scale:  3,
	Title:  "NES",
}

func init() {
	runtime.GOMAXPROCS(2)
	runtime.LockOSThread()
}

type Emulator struct {
	PlayerOneController ui.ControllerAdapter
	PlayerTwoController ui.ControllerAdapter

	Director *ui.Director
}

func NewEmulator(settings Settings, controllerOne ui.ControllerAdapter, controllerTwo ui.ControllerAdapter) (*Emulator, error) {
	log.Println(settings)
	// initialize audio
	portaudio.Initialize()
	defer portaudio.Terminate()

	audio := ui.NewAudio()
	if err := audio.Start(); err != nil {
		return nil, err
	}
	defer audio.Stop()

	// initialize glfw
	if err := glfw.Init(); err != nil {
		return nil, err
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(settings.Width*settings.Scale, settings.Height*settings.Scale, settings.Title, nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()

	// initialize gl
	if err := gl.Init(); err != nil {
		return nil, err
	}
	gl.Enable(gl.TEXTURE_2D)

	controllerOne.SetWindow(window)
	controllerTwo.SetWindow(window)

	d := ui.NewDirector(window, audio, controllerOne, controllerTwo)

	e := &Emulator{
		PlayerOneController: controllerOne,
		PlayerTwoController: controllerTwo,
		Director:            d,
	}

	e.Play("/Users/harrison/Downloads/Mario Bros. (World).nes")

	return e, nil
}

func (e *Emulator) Play(romPath string) {
	e.Director.Start([]string{romPath})
}

type Settings struct {
	Width  int
	Height int
	Scale  int
	Title  string
}
