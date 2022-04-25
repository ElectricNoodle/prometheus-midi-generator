package fractals

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

/*FractalType Type of fractal to draw */
type FractalType int

/*Values for above */
const (
	MandleBrot FractalType = 0
	JuliaSet   FractalType = 1
)

var lastTime = 0.0
var frameCount = 0

/*MandlebrotInfo Stores all the useful variables for the Mandelbrot Shader*/
type MandlebrotInfo struct {
	position         mgl32.Vec2
	zoom             float32
	rotation         float32
	rotationPivot    mgl32.Vec2
	colorModes       []int32
	colorOffsets     []float32
	maxIterations    float32
	exponentOne      float32
	exponentTwo      float32
	divideModifier   float32
	multiplyModifier float32
	escapeModifier   float32
}

/*FractalRenderer Defines a Fractal Renderer*/
type FractalRenderer struct {
	initialized    bool
	program        uint32
	vao            uint32
	fps            int
	vertexShader   uint32
	fragmentShader uint32
	activeFractal  *FractalType
	mandlebrotInfo MandlebrotInfo
	keysEnabled    bool
	keyPressMap    map[glfw.Key]bool
}

/*NewFractalRenderer Returns a new instance of FractalRenderer */
func NewFractalRenderer(logIn *logging.Logger) *FractalRenderer {

	log = logIn
	renderer := FractalRenderer{false, 0, 0, 0, 0, 0.0, nil,
		MandlebrotInfo{mgl32.Vec2{0.0, 0.0}, 1.5, 0.0, mgl32.Vec2{0.0, 0.0}, []int32{0, 2, 1}, []float32{0.0, 0.0, 0.0}, 20, 2, 2, 1.0, 1.0, 0.0}, true, make(map[glfw.Key]bool)}

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

	renderer.updateFractalShaderVariables()

	gl.BindVertexArray(renderer.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(renderSurface)/3))

	if renderer.keysEnabled {
		renderer.handleKeyPresses()
	}

	renderer.mandlebrotInfo.rotation = renderer.mandlebrotInfo.rotation + 0.008
	renderer.mandlebrotInfo.position = mgl32.Vec2{-0.8 + (float32)(math.Cos(time))/6, (float32)(math.Cos(time)) / 6}

	renderer.updateFPS(time)
	glfw.PollEvents()
}

func (renderer *FractalRenderer) updateFractalShaderVariables() {

	posOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("posOffset"+"\x00"))
	zoomOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("zoomOffset"+"\x00"))
	rotOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("rotOffset"+"\x00"))
	rotPivotLocation := gl.GetUniformLocation(renderer.program, gl.Str("rotPivot"+"\x00"))
	rModeLocation := gl.GetUniformLocation(renderer.program, gl.Str("rMode"+"\x00"))
	gModeLocation := gl.GetUniformLocation(renderer.program, gl.Str("gMode"+"\x00"))
	bModeLocation := gl.GetUniformLocation(renderer.program, gl.Str("bMode"+"\x00"))
	rOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("rMode"+"\x00"))
	gOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("gMode"+"\x00"))
	bOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("bMode"+"\x00"))
	iterOffsetLocation := gl.GetUniformLocation(renderer.program, gl.Str("maxIterations"+"\x00"))
	exponentOneLocation := gl.GetUniformLocation(renderer.program, gl.Str("exponentOne"+"\x00"))
	exponentTwoLocation := gl.GetUniformLocation(renderer.program, gl.Str("exponentTwo"+"\x00"))
	divModifierLocation := gl.GetUniformLocation(renderer.program, gl.Str("divModifier"+"\x00"))
	multModifierLocation := gl.GetUniformLocation(renderer.program, gl.Str("multModifier"+"\x00"))
	escapeModifierLocation := gl.GetUniformLocation(renderer.program, gl.Str("escapeModifier"+"\x00"))

	gl.UseProgram(renderer.program)
	gl.Uniform2fv(posOffsetLocation, 1, &renderer.mandlebrotInfo.position[0])
	gl.Uniform1f(zoomOffsetLocation, renderer.mandlebrotInfo.zoom)
	gl.Uniform1f(rotOffsetLocation, renderer.mandlebrotInfo.rotation)
	gl.Uniform2fv(rotPivotLocation, 1, &renderer.mandlebrotInfo.rotationPivot[0])
	gl.Uniform1i(rModeLocation, renderer.mandlebrotInfo.colorModes[0])
	gl.Uniform1i(gModeLocation, renderer.mandlebrotInfo.colorModes[1])
	gl.Uniform1i(bModeLocation, renderer.mandlebrotInfo.colorModes[2])
	gl.Uniform1f(rOffsetLocation, renderer.mandlebrotInfo.colorOffsets[0])
	gl.Uniform1f(bOffsetLocation, renderer.mandlebrotInfo.colorOffsets[1])
	gl.Uniform1f(gOffsetLocation, renderer.mandlebrotInfo.colorOffsets[2])
	gl.Uniform1f(iterOffsetLocation, renderer.mandlebrotInfo.maxIterations)
	gl.Uniform1f(exponentOneLocation, renderer.mandlebrotInfo.exponentOne)
	gl.Uniform1f(exponentTwoLocation, renderer.mandlebrotInfo.exponentTwo)
	gl.Uniform1f(divModifierLocation, renderer.mandlebrotInfo.divideModifier)
	gl.Uniform1f(multModifierLocation, renderer.mandlebrotInfo.multiplyModifier)
	gl.Uniform1f(escapeModifierLocation, renderer.mandlebrotInfo.escapeModifier)

}

func (renderer *FractalRenderer) handleKeyPresses() {

	if val, ok := renderer.keyPressMap[glfw.KeyUp]; ok {
		if val == true {
			renderer.mandlebrotInfo.position = renderer.mandlebrotInfo.position.Add(mgl32.Vec2{0, 0.01})
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyDown]; ok {
		if val == true {
			renderer.mandlebrotInfo.position = renderer.mandlebrotInfo.position.Add(mgl32.Vec2{0, -0.01})
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyLeft]; ok {
		if val == true {
			renderer.mandlebrotInfo.position = renderer.mandlebrotInfo.position.Add(mgl32.Vec2{-0.01, 0.0})
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyRight]; ok {
		if val == true {
			renderer.mandlebrotInfo.position = renderer.mandlebrotInfo.position.Add(mgl32.Vec2{0.01, 0.0})
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyA]; ok {
		if val == true {
			renderer.mandlebrotInfo.rotation += 0.01
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyZ]; ok {
		if val == true {
			renderer.mandlebrotInfo.rotation -= 0.01
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyS]; ok {
		if val == true {
			renderer.mandlebrotInfo.zoom -= 0.01
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyX]; ok {
		if val == true {
			renderer.mandlebrotInfo.zoom += 0.01
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyD]; ok {
		if val == true {
			renderer.mandlebrotInfo.maxIterations += 1
		}
	}

	if val, ok := renderer.keyPressMap[glfw.KeyC]; ok {
		if val == true {
			renderer.mandlebrotInfo.maxIterations -= 1
		}
	}

	renderer.mandlebrotInfo.rotationPivot = renderer.mandlebrotInfo.position

}

func (renderer *FractalRenderer) updateFPS(time float64) {
	if (time - lastTime) >= 1.0 {

		renderer.fps = frameCount
		frameCount = 0
		lastTime = time

	}
	frameCount++
}

/*KeyCallback Passed to platform so we get key events.
  Each event is a one shot, so we need to track which keys are held then move elsewhere in the render loop.
*/
func (renderer *FractalRenderer) KeyCallback(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	if action == glfw.Press {
		renderer.keyPressMap[key] = true
	}

	if action == glfw.Release {
		renderer.keyPressMap[key] = false
	}
}
