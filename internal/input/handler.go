package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Handler struct {
	prevKeyState    map[ebiten.Key]bool
	prevMouseState  map[ebiten.MouseButton]bool
}

func NewHandler() *Handler {
	return &Handler{
		prevKeyState:   make(map[ebiten.Key]bool),
		prevMouseState: make(map[ebiten.MouseButton]bool),
	}
}

func (h *Handler) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

func (h *Handler) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

func (h *Handler) IsKeyJustReleased(key ebiten.Key) bool {
	return inpututil.IsKeyJustReleased(key)
}

func (h *Handler) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

func (h *Handler) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

func (h *Handler) IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustReleased(button)
}

func (h *Handler) GetCursorPosition() (int, int) {
	return ebiten.CursorPosition()
}

func (h *Handler) GetMouseWheel() (float64, float64) {
	x, y := ebiten.Wheel()
	return x, y
}

func (h *Handler) IsAnyKeyPressed() bool {
	commonKeys := []ebiten.Key{
		ebiten.KeyEscape, ebiten.KeySpace, ebiten.KeyEnter,
		ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight,
		ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD,
		ebiten.KeyP, ebiten.KeyL,
	}
	
	for _, key := range commonKeys {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}
	
	return false
}