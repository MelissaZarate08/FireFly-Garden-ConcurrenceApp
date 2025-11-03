package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Handler maneja el input del usuario de forma eficiente
// Mantiene el estado anterior para detectar "just pressed" y "just released"
type Handler struct {
	prevKeyState    map[ebiten.Key]bool
	prevMouseState  map[ebiten.MouseButton]bool
}

// NewHandler crea un nuevo manejador de input
func NewHandler() *Handler {
	return &Handler{
		prevKeyState:   make(map[ebiten.Key]bool),
		prevMouseState: make(map[ebiten.MouseButton]bool),
	}
}

// IsKeyPressed retorna true si la tecla está presionada
func (h *Handler) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// IsKeyJustPressed retorna true solo en el frame que se presionó la tecla
func (h *Handler) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// IsKeyJustReleased retorna true solo en el frame que se soltó la tecla
func (h *Handler) IsKeyJustReleased(key ebiten.Key) bool {
	return inpututil.IsKeyJustReleased(key)
}

// IsMouseButtonPressed retorna true si el botón del mouse está presionado
func (h *Handler) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

// IsMouseButtonJustPressed retorna true solo en el frame que se presionó el botón
func (h *Handler) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

// IsMouseButtonJustReleased retorna true solo en el frame que se soltó el botón
func (h *Handler) IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustReleased(button)
}

// GetCursorPosition retorna la posición actual del cursor
func (h *Handler) GetCursorPosition() (int, int) {
	return ebiten.CursorPosition()
}

// GetMouseWheel retorna el desplazamiento de la rueda del mouse
func (h *Handler) GetMouseWheel() (float64, float64) {
	x, y := ebiten.Wheel()
	return x, y
}

// IsAnyKeyPressed retorna true si alguna tecla está presionada
func (h *Handler) IsAnyKeyPressed() bool {
	// Verificar teclas comunes
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