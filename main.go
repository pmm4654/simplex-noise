package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth int = 800
const winHeight int = 600

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1 color, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1 color, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)

		// this will get all of the colors between c1 and c2
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2)) // need to multiply by .2 because we are restricting it to .5
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5)) // just keeps the lerping in the the 0-1 range
		}
	}
	return result
}

func turbulence(x, y, frequency, lacunarity, gain float32, octaves int) float32 {
	var sum float32
	amplitude := float32(1)

	for i := 0; i < octaves; i++ {
		noise := snoise2(x*frequency, y*frequency) * amplitude
		if noise < 0 {
			noise = -1.0 * noise // avoid go's 64 bit absolute value function so we need no conversions
		}
		sum += noise
		frequency *= lacunarity
		amplitude *= gain
	}
	return sum
}

// fractal brownian motion
// lacunarity is the rate that you will be changing the frequency through each iteration
// gain is the rate that we will be changing this amplitude of the noise through each iteration
// octaves - how many iterations we are going to do
func fbm2(x float32, y float32, frequency float32, lacunarity float32, gain float32, octaves int) float32 {
	var sum float32
	amplitude := float32(1.0)
	for i := 0; i < octaves; i++ {
		// amplitude can make resulting numbers bigger or smaller - like how we scaled out the snoise function by dividing it by 100 earlier
		sum += snoise2(x*frequency, y*frequency) * amplitude
		frequency = frequency * lacunarity
		amplitude = amplitude * gain
	}
	return sum
}

// ensures the value is within the range you want
func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min float32, max float32, gradient []color, pixels []byte) {
	scale := 255.0 / (max - min) // scale that to the difference in range of your noise
	offset := min * scale        // have an offset to push that range to start at 0 and count up

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4 // pixel index
		pixels[p] = c.r
		pixels[p+1] = c.g
		pixels[p+2] = c.b
	}
}

func makeNoise(pixels []byte, frequency float32, lacunarity float32, gain float32, octaves int) {
	fmt.Println(fmt.Sprintf("Frequency: %f, Lacunarity: %f, gain: %f, octaves: %d", frequency, lacunarity, gain, octaves))
	noise := make([]float32, winHeight*winWidth)

	i := 0
	min := float32(9999.0)
	max := float32(-9999.0)
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			// noise[i] = fbm2(float32(x), float32(y), frequency, lacunarity, gain, octaves)
			noise[i] = turbulence(float32(x), float32(y), frequency, lacunarity, gain, octaves)
			if noise[i] < min {
				min = noise[i]
			} else if noise[i] > max {
				max = noise[i]
			}
			i++
		}
	}

	gradient := getDualGradient(color{0, 0, 175}, color{80, 160, 244}, color{12, 192, 75}, color{255, 255, 255})
	rescaleAndDraw(noise, min, max, gradient, pixels)
}

type color struct {
	r, g, b byte
}

func setPixel(x int, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4

	// index would be whateveryou y is, so say 1 and your x is 2, you would
	// want multiply by your width (3 in the example below) and add the x (2).
	// So you would multiply 1 * 3 + 2 and your pixel would be placed at 5 (x = 2,y = 1)
	//
	// 0 1 2
	// 3 4 [5]
	// 6 7 8
	//
	// Multiplying by 4 is because there are 4 bytes per pixel in the format sdl.PIXELFORMAT_ABGR8888
	// The 4 bytes are taken up by alpha, red, greed and blue

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func main() {

	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer texture.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)

	frequency := float32(.01)
	gain := float32(.2)
	lacunarity := float32(3.0)
	octaves := 3
	makeNoise(pixels, frequency, lacunarity, gain, octaves)
	keyState := sdl.GetKeyboardState()
	// Big Game Loop
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}

			mult := 1
			if keyState[sdl.SCANCODE_LSHIFT] != 0 || keyState[sdl.SCANCODE_RSHIFT] != 0 {
				mult = -1
			}
			if keyState[sdl.SCANCODE_O] != 0 {
				octaves = octaves + 1*mult
				makeNoise(pixels, frequency, lacunarity, gain, octaves)
			}

			if keyState[sdl.SCANCODE_F] != 0 {
				frequency = frequency + .001*float32(mult)
				makeNoise(pixels, frequency, lacunarity, gain, octaves)
			}

			if keyState[sdl.SCANCODE_G] != 0 {
				gain = gain + 0.1*float32(mult)
				makeNoise(pixels, frequency, lacunarity, gain, octaves)
			}

			if keyState[sdl.SCANCODE_L] != 0 {
				lacunarity = lacunarity + 0.1*float32(mult)
				makeNoise(pixels, frequency, lacunarity, gain, octaves)
			}

			if keyState[sdl.SCANCODE_R] != 0 {
				frequency = float32(.01)
				gain = float32(.2)
				lacunarity = float32(3.0)
				octaves = 3
				makeNoise(pixels, frequency, lacunarity, gain, octaves)
			}

			texture.Update(nil, pixels, winWidth*4)
			renderer.Copy(texture, nil, nil)
			renderer.Present()

			sdl.Delay(16)
		}
	}
}

// noise
//
// simples form of noise is to pick random numbers on a line
// Noise is a randomness to get a cool effect
//
// Coherent noise
// 2 important properties
// if an input is close to another, then the output will also be close
// if the inputs are far away from each other, then the outputs will be random and unpredictable
//
// How do we make coherent noise?
// 2 broad categories
//
// Value noise - simpler
//   1D example: value noise is pick regular intervals and at each point you come up with a random number
//   For the spaces in between the random intervals you do a linear interpolation (center between 2 points)
//
// Gradient Noise
//   Random things you pick at random intervals
//   In the 1D case rather than picking a random point, you pick a random slope
//   and then instead of linear interpolation of the points, you do a linear interpolation of the slopes
//   and then you get a curve from one slope to the next
//
//   2 main types:
//     Perlin noise
//     Simplex Noise
//     (Open Simplex and other variants, too)
//
// What do we do with it?  THe idea is procedural generation.  Creating various types of natural effects.  Cloud textures, geometry to look like hills/mountains, etc.
//
// 1 technique is called Fractal noise (fractal Brownian motion)
//   (another 1D example) Say you had a nice curved line and you change the
//   frequency of the curve you are producing so you ahve the same curve on a smaller scale and you can layer them together and end up with a very natural looking curve
//
//
