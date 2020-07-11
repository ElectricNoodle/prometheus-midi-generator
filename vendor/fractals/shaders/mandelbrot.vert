#version 410

layout(location = 0) in vec3 vertexPosition_modelspace;
uniform float uTime;

// Used for moving/zooming/rotating fractal.
uniform vec2 posOffset = vec2(-1.2,0);
uniform float zoomOffset = 1.0;
uniform float rotOffset = 0.0;

// Used to keep track of coloring modes per channel.
uniform int rMode = 0;
uniform int gMode = 2;
uniform int bMode = 1;

// Used for adding an offset to each color channel.
uniform float rOffset = 0.0;
uniform float gOffset = 0.6;
uniform float bOffset = 0.1;

// Overall max number of iterations
uniform float maxIterations = 20;

// What power to use in the Mandelbrot equation.
int exponentOne = 2;
int exponentTwo = 2;

// Applied to the return value of the Mandelbrot equation. 
uniform float divModifier = 1.0;
uniform float multModifier = 1.0;

// Applied to the conditional in the Mandelrot equation.
uniform float escapeModifier = 0.0;

// Values that stay constant for the whole mesh.
uniform mat4 uMVP;
out vec2 iCoord;

void main() {

    gl_Position = uMVP * vec4(vertexPosition_modelspace,1);
    iCoord = vertexPosition_modelspace.xy;

}



