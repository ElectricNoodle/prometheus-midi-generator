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
        if(length(z.x)>maxIterations + (cos(u_time)*19) ) return i/maxIterations; // dd this to maxIterations for fun + (cos(u_time)*19.3)

    }
    return maxIterations;
}

vec2 rotateUV(vec2 uv, vec2 pivot, float rotation) {
    float sine = sin(rotation);
    float cosine = cos(rotation);

    uv -= pivot;
    uv.x = uv.x * cosine - uv.y * sine;
    uv.y = uv.x * sine + uv.y * cosine;
    uv += pivot;

    return uv;
}

vec2 rotate(vec2 uv, vec2 pivot, float angle) { 

    float s = sin(angle);
    float c = cos(angle);

    mat2 rotationMatrix = mat2( c, s,
                            -s,  c);

    return vec2(rotationMatrix * (uv - pivot) + pivot);
}

void main() {

    vec2 pivot = vec2( 0.0, 0.0);
    vec2 movingCoord = vec2(coord.x+sin(u_time)/2 - 0.5,coord.y);
    vec2 rotationCoord = rotate(movingCoord, pivot, u_time/4);

    frag_colour = vec4(clamp(iterateMandelbrot(rotationCoord)/(abs(tan(u_time))),0,0.25), iterateMandelbrot(rotationCoord), iterateMandelbrot(rotationCoord)/2, 1.0);
    maxIterations = maxIterations +  (cos(u_time) * 5);

}
