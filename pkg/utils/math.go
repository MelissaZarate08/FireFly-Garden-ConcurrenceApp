package utils

import (
	"math"
	"math/rand"
)

type Vector2D struct {
	X float64
	Y float64
}

func NewVector2D(x, y float64) Vector2D {
	return Vector2D{X: x, Y: y}
}

func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

func (v Vector2D) Mul(scalar float64) Vector2D {
	return Vector2D{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2D) Normalize() Vector2D {
	mag := v.Magnitude()
	if mag == 0 {
		return Vector2D{X: 0, Y: 0}
	}
	return Vector2D{
		X: v.X / mag,
		Y: v.Y / mag,
	}
}

func Distance(v1, v2 Vector2D) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandomVector2D(minX, maxX, minY, maxY float64) Vector2D {
	return Vector2D{
		X: RandomFloat(minX, maxX),
		Y: RandomFloat(minY, maxY),
	}
}

func RandomUnitVector() Vector2D {
	angle := rand.Float64() * 2 * math.Pi
	return Vector2D{
		X: math.Cos(angle),
		Y: math.Sin(angle),
	}
}

func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func WrapAround(pos Vector2D, width, height float64) Vector2D {
	x := pos.X
	y := pos.Y

	if x < 0 {
		x = width
	} else if x > width {
		x = 0
	}

	if y < 0 {
		y = height
	} else if y > height {
		y = 0
	}

	return Vector2D{X: x, Y: y}
}