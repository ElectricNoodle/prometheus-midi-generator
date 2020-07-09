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