package utils

import (
	"math"
	"math/rand"
)

// Vector2D representa un vector 2D con coordenadas X e Y
type Vector2D struct {
	X float64
	Y float64
}

// NewVector2D crea un nuevo vector 2D
func NewVector2D(x, y float64) Vector2D {
	return Vector2D{X: x, Y: y}
}

// Add suma dos vectores
func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub resta dos vectores
func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Mul multiplica el vector por un escalar
func (v Vector2D) Mul(scalar float64) Vector2D {
	return Vector2D{
		X: v.X * scalar,
		Y: v.Y * scalar,
	}
}

// Magnitude devuelve la magnitud del vector
func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize normaliza el vector (magnitud = 1)
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

// Distance calcula la distancia entre dos puntos
func Distance(v1, v2 Vector2D) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// RandomFloat devuelve un float aleatorio entre min y max
func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// RandomVector2D devuelve un vector aleatorio dentro de los límites
func RandomVector2D(minX, maxX, minY, maxY float64) Vector2D {
	return Vector2D{
		X: RandomFloat(minX, maxX),
		Y: RandomFloat(minY, maxY),
	}
}

// RandomUnitVector devuelve un vector unitario en dirección aleatoria
func RandomUnitVector() Vector2D {
	angle := rand.Float64() * 2 * math.Pi
	return Vector2D{
		X: math.Cos(angle),
		Y: math.Sin(angle),
	}
}

// Clamp limita un valor entre min y max
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Lerp interpola linealmente entre a y b por t (0-1)
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// WrapAround envuelve una posición dentro de los límites (efecto toroidal)
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