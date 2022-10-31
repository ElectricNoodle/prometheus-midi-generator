package graph

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"

	"github.com/ElectricNoodle/prometheus-midi-generator/logging"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var log *logging.Logger

var (
	renderSurface = []float32{
		-0.95, 0.75, 0,
		-0.95, -0.75, 0,
		0.95, -0.75, 0,

		-0.95, 0.75, 0,
		0.95, 0.75, 0,
		0.95, -0.75, 0,
	}
)

type GraphRenderer struct {
	initialized bool
	program        uint32
	vao            uint32
	fps            int
	vertexShader   uint32
	fragmentShader uint32
}

var lastTime = 0.0
var frameCount = 0

func NewGraphRenderer(logIn *logging.Logger) *GraphRenderer {
	log = logIn

}

/*Init FractCalled to setup the OpenGL stuff when we're about to go into the loop in the GUI.*/
func (renderer *FractalRenderer) Init() {

	renderer.initOpenGL()
	renderer.makeVao(renderSurface)

	renderer.initialized = true
}

/*loadShader Loads in a shader from a file.*/
func (renderer *GraphRenderer) loadShader(path string) string {

	shaderBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	shaderString := string(shaderBytes)

	return shaderString
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func (renderer *GraphRenderer) initOpenGL() {

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Printf("OpenGL version %s\n", version)

	vertexShaderSource := renderer.loadShader("graph/shaders/graph.vert") + "\x00"
	vertexShader, err := renderer.compileShader(vertexShaderSource, gl.VERTEX_SHADER)

	if err != nil {
		panic(err)
	}

	fragmentShaderSource := renderer.loadShader("graph/shaders/graph.frag") + "\x00"
	fragmentShader, err := renderer.compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)

	if err != nil {
		panic(err)
	}

	renderer.vertexShader = vertexShader
	renderer.fragmentShader = fragmentShader

	prog := gl.CreateProgram()

	gl.AttachShader(prog, renderer.vertexShader)
	gl.AttachShader(prog, renderer.fragmentShader)

	gl.LinkProgram(prog)
	renderer.program = prog

}