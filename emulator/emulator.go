package emulator

import (
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/gordonklaus/portaudio"
	"github.com/paked/nes/nes"
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

	Settings Settings
	savePath string
}

func NewEmulator(settings Settings, controllerOne ui.ControllerAdapter, controllerTwo ui.ControllerAdapter, savePath string) (*Emulator, error) {
	e := &Emulator{
		PlayerOneController: controllerOne,
		PlayerTwoController: controllerTwo,
		Settings:            settings,

		savePath: savePath,
	}

	return e, nil
}

func (e *Emulator) Play(romPath string) error {
	// initialize audio
	portaudio.Initialize()
	defer portaudio.Terminate()

	audio := ui.NewAudio()
	if err := audio.Start(); err != nil {
		return err
	}
	defer audio.Stop()

	// initialize glfw
	if err := glfw.Init(); err != nil {
		return err
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(e.Settings.Width*e.Settings.Scale, e.Settings.Height*e.Settings.Scale, e.Settings.Title, nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	// initialize gl
	if err := gl.Init(); err != nil {
		return err
	}
	gl.Enable(gl.TEXTURE_2D)

	e.PlayerOneController.SetWindow(window)
	e.PlayerTwoController.SetWindow(window)

	e.Director = ui.NewDirector(window, audio, e.PlayerOneController, e.PlayerTwoController)

	go func() {
		time.Sleep(time.Second)

		e.LoadState(e.savePath)
	}()

	e.Director.Start([]string{romPath})

	return nil
}

func (e *Emulator) SaveState(path string) error {
	c := e.console()

	return c.SaveState(path)
}

func (e *Emulator) LoadState(path string) error {
	c := e.console()

	return c.LoadState(path)
}

func (e *Emulator) console() *nes.Console {
	return e.Director.Console()
}

type Settings struct {
	Width  int
	Height int
	Scale  int
	Title  string
}
