package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/core"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

// Renderer contiene funciones puras de renderizado
// No mantiene estado, solo transforma datos en píxeles
type Renderer struct{}

// NewRenderer crea un nuevo renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// DrawBackground dibuja el fondo nocturno con gradiente
func (r *Renderer) DrawBackground(screen *ebiten.Image) {
	screen.Fill(utils.ArrayToRGBA(config.BackgroundColor))
	
	// Efecto de gradiente sutil de arriba hacia abajo
	width := float32(config.ScreenWidth)
	height := float32(config.ScreenHeight)
	
	for i := 0; i < 3; i++ {
		y := float32(i) * height / 3
		alpha := uint8(20 - i*5)
		clr := color.RGBA{R: 20, G: 25, B: 50, A: alpha}
		vector.DrawFilledRect(screen, 0, y, width, height/3, clr, false)
	}
}

// DrawFirefly dibuja una luciérnaga con efecto de brillo
func (r *Renderer) DrawFirefly(screen *ebiten.Image, state core.FireflyState) {
	x := float32(state.Position.X)
	y := float32(state.Position.Y)
	
	// Interpolar color según brillo
	clr := utils.LerpColor(config.FireflyColorDim, config.FireflyColorFull, state.Brightness)
	
	// Halo externo (suavizado y con gradiente)
if state.Brightness > 0.15 {
	haloRadius := float32(config.FireflySize * 2.8 * state.Brightness)
	haloColor := utils.WithAlpha(clr, uint8(float64(clr.A)*0.28))
	vector.DrawFilledCircle(screen, x, y, haloRadius, haloColor, false)
}

// Halo medio
if state.Brightness > 0.1 {
	midRadius := float32(config.FireflySize * 1.6 * (0.7 + 0.6*state.Brightness))
	midColor := utils.WithAlpha(clr, uint8(float64(clr.A)*0.55))
	vector.DrawFilledCircle(screen, x, y, midRadius, midColor, false)
}
	
	// Dibujar núcleo brillante
	coreRadius := float32(config.FireflySize * state.Brightness)
	if coreRadius < 2 {
		coreRadius = 2
	}
	vector.DrawFilledCircle(screen, x, y, coreRadius, clr, false)
	
	// Centro ultra brillante
	if state.Brightness > 0.7 {
		centerColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(255 * state.Brightness)}
		vector.DrawFilledCircle(screen, x, y, coreRadius*0.5, centerColor, false)
	}
}

// DrawLantern dibuja un farol con efecto de pulso
func (r *Renderer) DrawLantern(screen *ebiten.Image, lantern *core.Lantern) {
	x := float32(lantern.Position.X)
	y := float32(lantern.Position.Y)
	intensity := lantern.GetIntensity()
	
	// Color base del farol
	baseColor := utils.ArrayToRGBA(config.LanternColor)
	
	// Dibujar aura de influencia (círculo grande transparente)
	auraRadius := float32(lantern.Radius)
	auraColor := utils.WithAlpha(baseColor, uint8(30*intensity))
	vector.DrawFilledCircle(screen, x, y, auraRadius, auraColor, false)
	
	// Dibujar anillos pulsantes
	for i := 0; i < 3; i++ {
		ringRadius := auraRadius * (0.3 + float32(i)*0.25)
		ringAlpha := uint8(50 * intensity * (1 - float64(i)*0.3))
		ringColor := utils.WithAlpha(baseColor, ringAlpha)
		vector.StrokeCircle(screen, x, y, ringRadius, 2, ringColor, false)
	}
	
	// Dibujar núcleo del farol
	coreRadius := float32(config.LanternSize)
	coreColor := utils.Brighten(baseColor, 0.3)
	vector.DrawFilledCircle(screen, x, y, coreRadius, coreColor, false)
	
	// Centro brillante
	centerRadius := coreRadius * 0.6
	centerColor := color.RGBA{R: 255, G: 240, B: 200, A: uint8(255 * intensity)}
	vector.DrawFilledCircle(screen, x, y, centerRadius, centerColor, false)
	
	// Punto ultra brillante central
	vector.DrawFilledCircle(screen, x, y, centerRadius*0.4, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
}

// DrawWind dibuja indicadores visuales del viento
func (r *Renderer) DrawWind(screen *ebiten.Image, wind *core.Wind) {
	force := wind.GetForce()
	
	// Dibujar partículas de viento en varias posiciones
	particleCount := 12
	particleColor := utils.ArrayToRGBA(config.WindColor)
	
	for i := 0; i < particleCount; i++ {
		// Posición inicial aleatoria pero determinística
		startX := float32(i * config.ScreenWidth / particleCount)
		startY := float32((i*137) % config.ScreenHeight) // Patrón pseudo-aleatorio
		
		// Línea que indica dirección del viento
		endX := startX + float32(force.X)*30
		endY := startY + float32(force.Y)*30
		
		// Dibujar línea de viento
		vector.StrokeLine(screen, startX, startY, endX, endY, 2, particleColor, false)
		
		// Dibujar punta de flecha
		r.drawArrowHead(screen, startX, startY, endX, endY, particleColor)
	}
}

// drawArrowHead dibuja la punta de una flecha
func (r *Renderer) drawArrowHead(screen *ebiten.Image, x1, y1, x2, y2 float32, clr color.RGBA) {
	// Calcular ángulo de la línea
	angle := math.Atan2(float64(y2-y1), float64(x2-x1))
	arrowSize := float32(8.0)
	
	// Puntas de la flecha (dos líneas)
	angle1 := angle + math.Pi*0.75
	angle2 := angle - math.Pi*0.75
	
	tipX1 := x2 + arrowSize*float32(math.Cos(angle1))
	tipY1 := y2 + arrowSize*float32(math.Sin(angle1))
	
	tipX2 := x2 + arrowSize*float32(math.Cos(angle2))
	tipY2 := y2 + arrowSize*float32(math.Sin(angle2))
	
	vector.StrokeLine(screen, x2, y2, tipX1, tipY1, 2, clr, false)
	vector.StrokeLine(screen, x2, y2, tipX2, tipY2, 2, clr, false)
}

// DrawAttractionPoint dibuja el punto de atracción cuando el jugador hace click
func (r *Renderer) DrawAttractionPoint(screen *ebiten.Image, point utils.Vector2D, pulse float64) {
	x := float32(point.X)
	y := float32(point.Y)
	
	// Efecto de pulso
	baseRadius := float32(15 + pulse*10)
	clr := color.RGBA{R: 255, G: 255, B: 100, A: uint8(150 * (1 - pulse))}
	
	// Círculos concéntricos pulsantes
	vector.StrokeCircle(screen, x, y, baseRadius, 3, clr, false)
	vector.StrokeCircle(screen, x, y, baseRadius*0.6, 2, clr, false)
	
	// Cruz en el centro
	crossSize := float32(8)
	vector.StrokeLine(screen, x-crossSize, y, x+crossSize, y, 2, clr, false)
	vector.StrokeLine(screen, x, y-crossSize, x, y+crossSize, 2, clr, false)
}

// DrawGrid dibuja una grilla de referencia (útil para debug)
func (r *Renderer) DrawGrid(screen *ebiten.Image, cellSize int) {
	gridColor := color.RGBA{R: 50, G: 50, B: 80, A: 50}
	
	// Líneas verticales
	for x := 0; x < config.ScreenWidth; x += cellSize {
		vector.StrokeLine(
			screen,
			float32(x), 0,
			float32(x), float32(config.ScreenHeight),
			1, gridColor, false,
		)
	}
	
	// Líneas horizontales
	for y := 0; y < config.ScreenHeight; y += cellSize {
		vector.StrokeLine(
			screen,
			0, float32(y),
			float32(config.ScreenWidth), float32(y),
			1, gridColor, false,
		)
	}
}