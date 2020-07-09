package fractals

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
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
	layout(location = 0) in vec3 vertexPosition_modelspace;
	uniform float u_time;
	// Values that stay constant for the whole mesh.
	uniform mat4 u_mvp;
	out vec2 coord;

	void main() {

		gl_Position =  u_mvp * vec4(vertexPosition_modelspace,1);
		coord = vertexPosition_modelspace.xy;

	}
` + "\x00"

	fragmentShaderSource = `
	#version 410

	precision highp float;

	uniform float u_time;
	float maxIterations = 20;
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

	float mandelbrot( in vec2 c )
	{
	
		const float B = 256.0;
		float l = 0.0;
		vec2 z  = vec2(0.0);
		for( int i=0; i<maxIterations; i++ )
		{
			z = vec2( z.x*z.x - z.y*z.y, 2.0*z.x*z.y ) + c;
			if( dot(z,z)>(B*B) ) break;
			l += 1.0;
		}
	
		if( l>maxIterations ) return 0.0;

		// equivalent optimized smooth interation count
		float sl = l - log2(log2(dot(z,z))) + 4.0;
	
		float al = smoothstep( -0.1, 0.0, cos(0.5*6.2831*(52/100) ) );
		l = mix( l, sl, al );
	
		return l;
	}
	void main() {
		
		frag_colour = vec4(clamp(iterateMandelbrot(coord)/(abs(tan(u_time))),0,0.3), iterateMandelbrot(coord), iterateMandelbrot(coord)/2, 1.0);


		//vec3 col;
		//float l = mandelbrot(coord);
		//col += 0.5 + 0.5*cos( 3.0 + l*0.15 + vec3(0.0,0.6,1.0));
		//frag_colour = vec4( col, 1.0 );
	}
	` + "\x00"
)

/*FractalRenderer Defines a Fractal Renderer*/
type FractalRenderer struct {
	program     uint32
	vao         uint32
	initialized bool
}

/*NewFractalRenderer Returns a new instance of FractalRenderer */
func NewFractalRenderer() *FractalRenderer {

	renderer := FractalRenderer{0, 0, false}

	return &renderer
}

/*Init FractCalled to setup the OpenGL stuff when we're about to go into the loop in the GUI.*/
func (renderer *FractalRenderer) Init() {

	renderer.initOpenGL()
	renderer.makeVao(square)

	renderer.initialized = true
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
	cameraX := (float32)(1.5) //(float32)(math.Sin(time) * 0.5)
	cameraY := (float32)(0)   //(float32)(math.Cos(time) * radius)
	cameraZ := (float32)(5.0)

	view := mgl32.LookAt(
		cameraX, cameraY, cameraZ,
		1.5, 0, 0,
		0, 1, 0)

	gl.UseProgram(renderer.program)

	model := mgl32.Ident4()
	scale := mgl32.Scale3D(2.0, 2.0, 2.0)

	//rotationX := mgl32.HomogRotate3DX(float32(time))
	//rotationY := mgl32.HomogRotate3DY(float32(time / 2))
	//rotationZ := mgl32.HomogRotate3DZ(float32(time / 4))

	timeLocation := gl.GetUniformLocation(renderer.program, gl.Str("u_time"+"\x00"))
	modelViewProjection := projection.Mul4(view).Mul4(model).Mul4(scale) //.Mul4(rotationY).Mul4(rotationX).Mul4(rotationZ)
	shaderMvp := gl.GetUniformLocation(renderer.program, gl.Str("u_mvp"+"\x00"))

	gl.UseProgram(renderer.program)
	gl.Uniform1f(timeLocation, float32(time))
	gl.UniformMatrix4fv(shaderMvp, 1, false, &modelViewProjection[0])

	gl.BindVertexArray(renderer.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))

	glfw.PollEvents()
}
