package graph

import (
	"fmt"
	"io/ioutil"
//	"math"
	"strings"

	"github.com/ElectricNoodle/prometheus-midi-generator/logging"

	"github.com/go-gl/gl/v3.2-core/gl"
//	"github.com/go-gl/glfw/v3.2/glfw"
//	"github.com/go-gl/mathgl/mgl32"
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
	renderer := GraphRenderer{
		initialized:	false,
		program:		0,
		vao: 			0,
		fps:			0,
		vertexShader:	0,
		fragmentShader: 0,
	}
	return &renderer
}

// makeVao initializes and returns a vertex array from the points provided.
func (renderer *GraphRenderer) makeVao(points []float32) {

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	renderer.vao = vao
}

/*Init Called to setup the OpenGL stuff when we're about to go into the loop in the GUI.*/
func (renderer *GraphRenderer) Init() {

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

func (renderer *GraphRenderer) compileShader(source string, shaderType uint32) (uint32, error) {

	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)

	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
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

/*Render Called to draw the fractal */
func (renderer *GraphRenderer) Render(displaySize [2]float32, framebufferSize [2]float32) {

}
