package render

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/core"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

//go:embed assets/Go-Regular.ttf
var goTTF []byte

// UIRenderer maneja el renderizado de elementos de interfaz
type UIRenderer struct {
	fontFace  text.Face
	largeFace text.Face
}

// NewUIRenderer crea un nuevo renderizador de UI
func NewUIRenderer() *UIRenderer {
	// Crear fuente básica de Go (tamaño normal) usando TTF embebido
	if len(goTTF) == 0 {
		log.Fatalf("fuente embebida vacía: asegúrate de que internal/render/assets/Go-Regular.ttf exista")
	}

	src, err := text.NewGoTextFaceSource(bytes.NewReader(goTTF))
	if err != nil {
		log.Fatalf("Error al crear fuente (normal): %v", err)
	}
	fontFace := &text.GoTextFace{
		Source: src,
		Size:   16,
	}

	// Crear fuente grande para mensajes
	largeSrc, err := text.NewGoTextFaceSource(bytes.NewReader(goTTF))
	if err != nil {
		log.Fatalf("Error al crear fuente (grande): %v", err)
	}
	largeFace := &text.GoTextFace{
		Source: largeSrc,
		Size:   48,
	}

	return &UIRenderer{
		fontFace:  fontFace,
		largeFace: largeFace,
	}
}

// DrawHUD dibuja el HUD principal con información del juego
func (u *UIRenderer) DrawHUD(screen *ebiten.Image, fireflyCount, lanternCount int, wind *core.Wind, fps float64, isPaused bool) {
	padding := 10.0
	lineHeight := 22.0
	y := padding

	// Panel semi-transparente de fondo
	panelHeight := float32(lineHeight * 7)
	panelColor := color.RGBA{R: 0, G: 0, B: 0, A: 150}
	vector.DrawFilledRect(screen, float32(padding), float32(y), 300, panelHeight, panelColor, false)

	// Título
	u.drawText(screen, "🌙 JARDÍN DE LUCIÉRNAGAS", padding+10, y+4, color.RGBA{R: 255, G: 255, B: 200, A: 255})
	y += lineHeight

	// Separador
	vector.StrokeLine(screen, float32(padding+10), float32(y), float32(padding+280), float32(y), 1, color.RGBA{R: 100, G: 100, B: 100, A: 120}, false)
	y += lineHeight * 0.5

	// Estadísticas
	textColor := utils.ArrayToRGBA(config.UITextColor)

	u.drawText(screen, fmt.Sprintf("Luciérnagas: %d / %d", fireflyCount, config.MaxFireflies), padding+10, y, textColor)
	y += lineHeight

	u.drawText(screen, fmt.Sprintf("Faroles: %d / %d", lanternCount, config.MaxLanterns), padding+10, y, textColor)
	y += lineHeight

	u.drawText(screen, fmt.Sprintf("Viento: %s", wind.GetDirectionName()), padding+10, y, textColor)
	y += lineHeight

	u.drawText(screen, fmt.Sprintf("Objetivo: %d", config.ObjectiveCount), padding+10, y, textColor)
	y += lineHeight

	u.drawText(screen, fmt.Sprintf("FPS: %.1f  Goroutines: %d", fps, runtime.NumGoroutine()), padding+10, y, textColor)
	y += lineHeight

	// Estadística de estados descartados por canal
	dropped := core.GetDroppedStates()
	u.drawText(screen, fmt.Sprintf("Descartados: %d", dropped), padding+10, y, color.RGBA{R: 240, G: 200, B: 120, A: 255})
	y += lineHeight

	// Estado de pausa
	if isPaused {
		pauseColor := color.RGBA{R: 255, G: 100, B: 100, A: 255}
		u.drawText(screen, "⏸ PAUSADO", padding+10, y, pauseColor)
	}
}

// DrawControls dibuja la guía de controles
func (u *UIRenderer) DrawControls(screen *ebiten.Image) {
	x := float64(config.ScreenWidth - 320)
	y := 10.0
	lineHeight := 22.0

	// Panel de fondo
	panelHeight := float32(lineHeight * 8)
	panelColor := color.RGBA{R: 0, G: 0, B: 0, A: 150}
	vector.DrawFilledRect(screen, float32(x), float32(y), 300, panelHeight, panelColor, false)

	// Título
	u.drawText(screen, "⌨️  CONTROLES", x+10, y+5, color.RGBA{R: 150, G: 200, B: 255, A: 255})
	y += lineHeight

	// Separador
	vector.StrokeLine(screen, float32(x+10), float32(y), float32(x+290), float32(y), 1, color.RGBA{R: 100, G: 100, B: 100, A: 100}, false)
	y += lineHeight * 0.5

	// Controles
	textColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	u.drawText(screen, "Click Izq: Atraer luciernagas", x+10, y, textColor)
	y += lineHeight

	u.drawText(screen, "L: Colocar farol (genera ráfaga)", x+10, y, textColor)
	y += lineHeight

	u.drawText(screen, "K: Generar ráfaga cerca del cursor", x+10, y, textColor)
	y += lineHeight

	u.drawText(screen, "W: Cambiar direccion viento", x+10, y, textColor)
	y += lineHeight

	u.drawText(screen, "P: Pausar/Reanudar", x+10, y, textColor)
	y += lineHeight

	u.drawText(screen, "ESC: Salir", x+10, y, textColor)
}

// DrawPauseOverlay dibuja un overlay cuando el juego está pausado
func (u *UIRenderer) DrawPauseOverlay(screen *ebiten.Image) {
	// Overlay semi-transparente
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	vector.DrawFilledRect(screen, 0, 0, float32(config.ScreenWidth), float32(config.ScreenHeight), overlayColor, false)

	// Texto grande de pausa
	centerX := float64(config.ScreenWidth / 2)
	centerY := float64(config.ScreenHeight / 2)

	pauseText := "⏸  JUEGO PAUSADO"

	// Medir texto para centrarlo
	textWidth := text.Advance(pauseText, u.largeFace)

	op := &text.DrawOptions{}
	op.GeoM.Translate(centerX-textWidth/2, centerY-24)
	op.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	text.Draw(screen, pauseText, u.largeFace, op)

	// Mensaje secundario
	u.drawTextCentered(screen, "Presiona P para continuar", centerY+40, color.RGBA{R: 200, G: 200, B: 200, A: 255})
}

// DrawButton dibuja un botón interactivo
func (u *UIRenderer) DrawButton(screen *ebiten.Image, x, y, width, height float32, label string, isHovered bool) {
	// Color del botón
	btnColor := color.RGBA{R: 50, G: 50, B: 80, A: 200}
	if isHovered {
		btnColor = color.RGBA{R: 70, G: 70, B: 120, A: 230}
	}

	// Dibujar botón
	vector.DrawFilledRect(screen, x, y, width, height, btnColor, false)

	// Borde
	borderColor := color.RGBA{R: 150, G: 150, B: 200, A: 255}
	vector.StrokeRect(screen, x, y, width, height, 2, borderColor, false)

	// Texto centrado
	textWidth := text.Advance(label, u.fontFace)
	textX := float64(x) + float64(width)/2 - textWidth/2
	textY := float64(y) + float64(height)/2 - 8

	u.drawText(screen, label, textX, textY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

// DrawObjectivePanel dibuja el panel de objetivos del juego
func (u *UIRenderer) DrawObjectivePanel(screen *ebiten.Image, fireflyCount int) {
	x := float64(config.ScreenWidth/2 - 150)
	y := float64(config.ScreenHeight - 100)
	width := float32(300)
	height := float32(80)

	// Panel de fondo
	panelColor := color.RGBA{R: 20, G: 20, B: 40, A: 200}
	vector.DrawFilledRect(screen, float32(x), float32(y), width, height, panelColor, false)

	// Borde
	borderColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	vector.StrokeRect(screen, float32(x), float32(y), width, height, 2, borderColor, false)

	// Título
	u.drawTextCentered(screen, "🎯 OBJETIVO", y+15, color.RGBA{R: 255, G: 255, B: 150, A: 255})

	// Progreso
	objective := config.ObjectiveCount
	progress := float64(fireflyCount) / float64(objective)
	if progress > 1.0 {
		progress = 1.0
	}

	progressText := fmt.Sprintf("Mantén %d+ luciérnagas", objective)
	u.drawTextCentered(screen, progressText, y+40, color.RGBA{R: 200, G: 200, B: 200, A: 255})

	// Barra de progreso
	barX := float32(x + 30)
	barY := float32(y + 55)
	barWidth := width - 60
	barHeight := float32(15)

	// Fondo de barra
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{R: 50, G: 50, B: 50, A: 255}, false)

	// Progreso de barra
	progressWidth := barWidth * float32(progress)
	progressColor := color.RGBA{R: 100, G: 255, B: 100, A: 255}
	if progress < 0.5 {
		progressColor = color.RGBA{R: 255, G: 200, B: 100, A: 255}
	}
	if progress < 0.25 {
		progressColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
	}

	vector.DrawFilledRect(screen, barX, barY, progressWidth, barHeight, progressColor, false)

	// Borde de barra
	vector.StrokeRect(screen, barX, barY, barWidth, barHeight, 1, color.RGBA{R: 150, G: 150, B: 150, A: 255}, false)
}

// drawText dibuja texto en la posición especificada
func (u *UIRenderer) drawText(screen *ebiten.Image, txt string, x, y float64, clr color.RGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, txt, u.fontFace, op)
}

// drawTextCentered dibuja texto centrado horizontalmente
func (u *UIRenderer) drawTextCentered(screen *ebiten.Image, txt string, y float64, clr color.RGBA) {
	textWidth := text.Advance(txt, u.fontFace)
	x := float64(config.ScreenWidth)/2 - textWidth/2
	u.drawText(screen, txt, x, y, clr)
}