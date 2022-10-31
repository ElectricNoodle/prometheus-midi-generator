#version 410
layout (location = 0) in vec3 vertexPosition_modelspace; // the position variable has attribute position 0
  

out vec4 vertexColor; // specify a color output to the fragment shader

void main()
{
    gl_Position = vec4(vertexPosition_modelspace, 1.0); // see how we directly give a vec3 to vec4's constructor
    vertexColor = vec4(0.5, 0.0, 0.0, 1.0); // set the output variable to a dark-red color
}