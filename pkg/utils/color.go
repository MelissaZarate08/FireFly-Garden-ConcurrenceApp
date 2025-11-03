package utils

import (
	"image/color"
)

func LerpColor(c1, c2 [4]uint8, t float64) color.RGBA {
	t = Clamp(t, 0, 1)
	
	r := uint8(Lerp(float64(c1[0]), float64(c2[0]), t))
	g := uint8(Lerp(float64(c1[1]), float64(c2[1]), t))
	b := uint8(Lerp(float64(c1[2]), float64(c2[2]), t))
	a := uint8(Lerp(float64(c1[3]), float64(c2[3]), t))
	
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func ArrayToRGBA(arr [4]uint8) color.RGBA {
	return color.RGBA{R: arr[0], G: arr[1], B: arr[2], A: arr[3]}
}

func WithAlpha(c color.RGBA, alpha uint8) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha}
}

func Brighten(c color.RGBA, factor float64) color.RGBA {
	factor = Clamp(factor, 0, 1)
	
	r := uint8(Clamp(float64(c.R)*(1+factor), 0, 255))
	g := uint8(Clamp(float64(c.G)*(1+factor), 0, 255))
	b := uint8(Clamp(float64(c.B)*(1+factor), 0, 255))
	
	return color.RGBA{R: r, G: g, B: b, A: c.A}
}