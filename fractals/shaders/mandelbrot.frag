#version 410

precision highp float;

in vec2 iCoord;


uniform float uTime;

// Used for moving/zooming/rotating fractal.
uniform vec2 posOffset = vec2(-1.2,0);
uniform float zoomOffset = 1.0;
uniform float rotOffset = 0.0;
uniform vec2 rotPivot = vec2(0,0);

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
uniform float exponentOne = 2;
uniform float exponentTwo = 2;

// Applied to the return value of the Mandelbrot equation. 
uniform float divModifier = 1.0;
uniform float multModifier = 1.0;

// Applied to the conditional in the Mandelrot equation.
uniform float escapeModifier = 0.0;

out vec4 fragColor;

vec2 squareImaginary(vec2 number){
    return vec2(
        pow(number.x,exponentOne)-pow(number.y,exponentTwo),
        exponentTwo*number.x*number.y
    );
}

float iterateMandelbrot(vec2 coord){
    vec2 z = vec2(0,0);
    for(int i=0;i<maxIterations;i++){

        z = squareImaginary(z) + coord;
        if(length(z.x)>maxIterations + (cos(uTime)*17) ) return i/maxIterations;

    }
    return maxIterations;
}

vec2 rotate(vec2 uv, vec2 pivot, float angle) { 

    float s = sin(angle);
    float c = cos(angle);
    mat2 rotationMatrix = mat2( c, s,
                            -s,  c);

    return vec2(rotationMatrix * (uv - pivot) + pivot);
}

// Returns different transforms of input num for colouring.
float getColor(float num, float offset, int mode) {
    
    switch(mode){
        case 0:
            return num;
        case 1:
            return num + offset;
        case 2:
            return (0.5 + 0.5 *cos(2.7 + num * 30.0 + offset)) ;
    }
    return 0.0;   
}

vec2 rotateScaleTranslate(vec2 position, float rotation, float zoom, vec2 posOffset, vec2 pivot) {

    float s = sin(rotation);
    float c = cos(rotation);

    mat3 translation = mat3(1,0, 0,
                            0,1, 0,
                            posOffset.x,posOffset.y,1);

    mat3 rotMat = mat3( c, s,0,
                        -s,c,0,
                         0,0,1);

    mat3 scaleMat = mat3(zoom,0,0,
                         0, zoom, 0,
                         0, 0, 1);



   // vec3 result =rotMat * scaleMat * translation  * vec3(position - pivot,1.0) + vec3(pivot, 1.0);
    vec3 result =rotMat * scaleMat * translation  *vec3(position,1.0);
    return vec2(result.x,result.y);
}

// Handles move/rotate/zoom of current coordinates.
// Calculates Mandelbrot value then uses it to set the colour depending on the mode.
void main() {

    vec2 coord = rotateScaleTranslate(iCoord, rotOffset, zoomOffset, posOffset,rotPivot);

//    coord = coord + posOffset;

    float mandleBrotValue = (iterateMandelbrot(coord) * multModifier / divModifier);
    
    fragColor = vec4( getColor(mandleBrotValue, rOffset, rMode),
                      getColor(mandleBrotValue, gOffset, gMode), 
                      getColor(mandleBrotValue, bOffset, bMode),
                      1.0);
}
