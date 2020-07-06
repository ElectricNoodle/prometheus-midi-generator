package fractals

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	square = []float32{
		-1.0, 1.0, 0,
		-1.0, -1.0, 0,
		1.0, -1.0, 0,

		-1.0, 1.0, 0,
		1.0, 1.0, 0,
		1.0, -1.0, 0,
	}

	vertexShaderSource = `
	#version 410
	attribute vec4 vPosition;
	layout (location = 0) in vec2 position;
	out vec2 coord;

	void main() {
		gl_Position = vec4(position, 0.0f, 0.8f);
		coord = position.xy;
	}
` + "\x00"

	fragmentShaderSource = `
	#version 410

	uniform float u_time;
	float maxIterations = 100;
	in vec2 coord;

	out vec4 frag_colour;

	vec2 squareImaginary(vec2 number){
		return vec2(
			pow(number.x,2)-pow(number.y,2),
			2*number.x*number.y
		);
	}
	
	float iterateMandelbrot(vec2 coord){
		vec2 z = vec2(0,0);
		for(int i=0;i<maxIterations;i++){
			z = squareImaginary(z) + coord;
			if(length(z.x)>2) return i/maxIterations;
		}
		return maxIterations;
	}

	void main() {

		frag_colour = vec4(clamp(iterateMandelbrot(coord)/(abs(tan(u_time))),0,0.3),iterateMandelbrot(coord),iterateMandelbrot(coord),1.0);

	}
	` + "\x00"
)

/*FractalRenderer Defines a Fractal Renderer*/
type FractalRenderer struct {
	program uint32
	vao     uint32
}

/*NewFractalRenderer Returns a new instance of FractalRenderer */
func NewFractalRenderer() *FractalRenderer {

	renderer := FractalRenderer{0, 0}
	renderer.initOpenGL()
	renderer.makeVao(square)

	return &renderer
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func (renderer *FractalRenderer) initOpenGL() {

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Printf("OpenGL version %s\n", version)

	vertexShader, err := renderer.compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := renderer.compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)

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
	gl.UseProgram(renderer.program)

	time := glfw.GetTime()

	timeLocation := gl.GetUniformLocation(renderer.program, gl.Str("u_time"+"\x00"))

	gl.UseProgram(renderer.program)
	gl.Uniform1f(timeLocation, float32(time))

	gl.BindVertexArray(renderer.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))

	glfw.PollEvents()
}
