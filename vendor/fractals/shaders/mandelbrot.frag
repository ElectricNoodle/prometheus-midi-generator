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
