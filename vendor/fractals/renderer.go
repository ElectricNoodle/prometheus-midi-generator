package fractals

import (
	"fmt"
	"io/ioutil"
	"logging"
	"strings"

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

/*FractalRenderer Defines a Fractal Renderer*/
type FractalRenderer struct {
	initialized    bool
	program        uint32
	vao            uint32
	vertexShader   uint32
	fragmentShader uint32
}

/*NewFractalRenderer Returns a new instance of FractalRenderer */
func NewFractalRenderer(logIn *logging.Logger) *FractalRenderer {

	log = logIn
	renderer := FractalRenderer{false, 0, 0, 0, 0}

	return &renderer
}

/*Init FractCalled to setup the OpenGL stuff when we're about to go into the loop in the GUI.*/
func (renderer *FractalRenderer) Init() {

	renderer.initOpenGL()
	renderer.makeVao(renderSurface)

	renderer.initialized = true
}

/*loadShader Loads in a shader from a file.*/
func (renderer *FractalRenderer) loadShader(path string) string {

	shaderBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	shaderString := string(shaderBytes)

	return shaderString
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func (renderer *FractalRenderer) initOpenGL() {

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Printf("OpenGL version %s\n", version)

	vertexShaderSource := renderer.loadShader("vendor/fractals/shaders/mandelbrot.vert") + "\x00"
	vertexShader, err := renderer.compileShader(vertexShaderSource, gl.VERTEX_SHADER)

	if err != nil {
		panic(err)
	}

	fragmentShaderSource := renderer.loadShader("vendor/fractals/shaders/mandelbrot.frag") + "\x00"
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

// makeVao initializes and returns a vertex array from the points provided.
func (renderer *FractalRenderer) makeVao(points []float32) {

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

func (renderer *FractalRenderer) compileShader(source string, shaderType uint32) (uint32, error) {

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

/*Render Called to draw the fractal */
func (renderer *FractalRenderer) Render(displaySize [2]float32, framebufferSize [2]float32) {

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	displayWidth, displayHeight := displaySize[0], displaySize[1]
	fbWidth, fbHeight := framebufferSize[0], framebufferSize[1]

	if (fbWidth <= 0) || (fbHeight <= 0) {
		return
	}

	time := glfw.GetTime()
	//	log.Println(time)
	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(displayWidth)/displayHeight, 0.1, 10.0)

	//	glm::vec3(4,3,3), // Camera is at (4,3,3), in World Space
	//	glm::vec3(0,0,0), // and looks at the origin
	//	glm::vec3(0,1,0)  // Head is up (set to 0,-1,0 to look upside-down)

	//radius := 5.0
	cameraX := (float32)(0.95) //(float32)(math.Sin(time) * 0.5)
	cameraY := (float32)(0)    //(float32)(math.Cos(time) * radius)
	cameraZ := (float32)(3.7)

	view := mgl32.LookAt(
		cameraX, cameraY, cameraZ,
		0.95, 0, 0,
		0, 1, 0)

	gl.UseProgram(renderer.program)

	model := mgl32.Ident4()
	scale := mgl32.Scale3D(2.0, 2.0, 2.0)

	timeLocation := gl.GetUniformLocation(renderer.program, gl.Str("uTime"+"\x00"))
	modelViewProjection := projection.Mul4(view).Mul4(model).Mul4(scale)
	shaderMvp := gl.GetUniformLocation(renderer.program, gl.Str("uMVP"+"\x00"))

	gl.UseProgram(renderer.program)
	gl.Uniform1f(timeLocation, float32(time))
	gl.UniformMatrix4fv(shaderMvp, 1, false, &modelViewProjection[0])

	gl.BindVertexArray(renderer.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(renderSurface)/3))

	glfw.PollEvents()
}

func (renderer *FractalRenderer) updateFPS(time float64) {

}

/*KeyCallback Passed to platform so we get key events. */
func (renderer *FractalRenderer) KeyCallback(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	log.Println("Debug:", key)
}
