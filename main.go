package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth int = 800
const winHeight int = 600

func rescaleAndDraw(noise []float32, min float32, max float32, pixels []byte) {
	scale := 255.0 / (max - min) // scale that to the difference in range of your noise
	offset := min * scale        // have an offset to push that range to start at 0 and count up

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		b := byte(noise[i])
		pixels[i*4] = b
		pixels[i*4+1] = b
		pixels[i*4+2] = b
	}
}

func makeNoise(pixels []byte) {
	noise := make([]float32, winHeight*winWidth)

	i := 0
	min := float32(9999.0)
	max := float32(-9999.0)
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			noise[i] = snoise2(float32(x)/100.0, float32(y)/100.0) // the smaller this number is the more zoomed out it appears to be
			if noise[i] < min {
				min = noise[i]
			} else if noise[i] > max {
				max = noise[i]
			}
			i++
		}
	}
	rescaleAndDraw(noise, min, max, pixels)
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

	makeNoise(pixels)

	texture.Update(nil, pixels, winWidth*4)
	renderer.Copy(texture, nil, nil)
	renderer.Present()
	// Big Game Loop

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
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
