package emulator

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/gordonklaus/portaudio"
	"github.com/paked/nes/ui"
)

const (
	Width  = 246
	Height = 240
	Scale  = 3
	Title  = "NES"
)

func init() {
	runtime.GOMAXPROCS(2)
	runtime.LockOSThread()
}

func Emulate(romPath string, playerOne ui.ControllerAdapter, playerTwo ui.ControllerAdapter) {
	// initialize audio
	portaudio.Initialize()
	defer portaudio.Terminate()

	audio := ui.NewAudio()
	if err := audio.Start(); err != nil {
		log.Fatalln(err)
	}
	defer audio.Stop()

	if err := glfw.Init(); err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(Width*Scale, Height*Scale, Title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()

	// initialize gl
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	gl.Enable(gl.TEXTURE_2D)

	director := ui.NewDirector(window, audio, playerOne, playerTwo)
	director.Start([]string{romPath})
}
