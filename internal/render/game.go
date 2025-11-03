package render

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/input"
	"github.com/yourusername/firefly-garden/internal/manager"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

// Game implementa la interfaz ebiten.Game
// Esta es la capa que conecta toda la lógica concurrente con Ebiten
type Game struct {
	manager           *manager.FireflyManager
	inputHandler      *input.Handler
	renderer          *Renderer
	uiRenderer        *UIRenderer
	gameState         int
	lastUpdateTime    time.Time
	attractionPulse   float64
	showAttraction    bool
	attractionPoint   utils.Vector2D
	fpsCounter        *FPSCounter

	// Nuevos campos para spawn del jugador
	lastPlayerSpawn     time.Time
	playerSpawnCooldown time.Duration
}

// FPSCounter calcula los FPS del juego
type FPSCounter struct {
	frames       int
	lastFPSTime  time.Time
	currentFPS   float64
}

// NewFPSCounter crea un nuevo contador de FPS
func NewFPSCounter() *FPSCounter {
	return &FPSCounter{
		frames:      0,
		lastFPSTime: time.Now(),
		currentFPS:  0,
	}
}

// Update actualiza el contador y retorna los FPS actuales
func (f *FPSCounter) Update() float64 {
	f.frames++

	now := time.Now()
	elapsed := now.Sub(f.lastFPSTime).Seconds()

	if elapsed >= 1.0 {
		f.currentFPS = float64(f.frames) / elapsed
		f.frames = 0
		f.lastFPSTime = now
	}

	return f.currentFPS
}

// NewGame crea una nueva instancia del juego
func NewGame() *Game {
	manager := manager.NewFireflyManager()
	inputHandler := input.NewHandler()

	game := &Game{
		manager:             manager,
		inputHandler:        inputHandler,
		renderer:            NewRenderer(),
		uiRenderer:          NewUIRenderer(),
		gameState:           config.GameStateRunning,
		lastUpdateTime:      time.Now(),
		fpsCounter:          NewFPSCounter(),
		lastPlayerSpawn:     time.Now().Add(-time.Hour),
		playerSpawnCooldown: time.Duration(config.PlayerSpawnCooldownSecs) * time.Second,
	}

	// Iniciar manager (arranca todas las goroutines)
	manager.Start()

	return game
}

// Update implementa ebiten.Game.Update
// Se llama TargetFPS veces por segundo
func (g *Game) Update() error {
	// Calcular delta time
	now := time.Now()
	dt := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Procesar input
	g.processInput(dt)

	// Actualizar lógica según estado del juego
	if g.gameState == config.GameStateRunning {
		g.updateGameLogic(dt)
	}

	// Actualizar contador de FPS
	g.fpsCounter.Update()

	return nil
}

// processInput procesa todos los inputs del usuario
func (g *Game) processInput(dt float64) {
	// Detectar tecla ESC para salir
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyEscape) {
		return // El juego se cerrará
	}

	// Detectar tecla P para pausar
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyP) {
		g.togglePause()
	}

	// Si está pausado, no procesar más inputs
	if g.gameState == config.GameStatePaused {
		return
	}

	// Detectar tecla L para crear farol
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyL) {
		mx, my := ebiten.CursorPosition()
		g.createLantern(float64(mx), float64(my))
	}

	// Detectar tecla W para cambiar viento
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyW) {
		g.changeWind()
	}

	// Detectar click izquierdo para atraer luciérnagas
	if g.inputHandler.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.setAttractionPoint(float64(mx), float64(my))
	}

	// Tecla K: Spawn burst cerca del cursor (feedback inmediato)
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyK) {
		// cooldown para evitar spam
		if time.Since(g.lastPlayerSpawn) >= g.playerSpawnCooldown {
			mx, my := ebiten.CursorPosition()
			// spawn burst via manager (no bloqueante)
			go g.manager.SpawnBurst(float64(mx), float64(my), config.SpawnBurstCount)
			g.lastPlayerSpawn = time.Now()
		}
	}

	// Si se suelta el botón, quitar atracción después de un tiempo
	if !g.inputHandler.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.showAttraction {
		g.attractionPulse += dt * 3
		if g.attractionPulse > 1.0 {
			g.clearAttractionPoint()
		}
	}
}

// updateGameLogic actualiza la lógica del juego
func (g *Game) updateGameLogic(dt float64) {
	// Actualizar faroles (animación de pulso)
	g.manager.UpdateLanterns(dt)

	// Actualizar pulso de atracción si está activo
	if g.showAttraction {
		g.attractionPulse += dt * 2
		if g.attractionPulse > 1.0 {
			g.attractionPulse = 0.0
		}
	}
}

// Draw implementa ebiten.Game.Draw
// Dibuja todos los elementos en pantalla
func (g *Game) Draw(screen *ebiten.Image) {
	// 1. Dibujar fondo
	g.renderer.DrawBackground(screen)

	// 2. Dibujar indicadores de viento
	g.renderer.DrawWind(screen, g.manager.GetWind())

	// 3. Dibujar faroles
	lanterns := g.manager.GetLanterns()
	for _, lantern := range lanterns {
		g.renderer.DrawLantern(screen, lantern)
	}

	// 4. Dibujar luciérnagas (obtener snapshot thread-safe)
	fireflyStates := g.manager.GetFireflyStates()
	for _, state := range fireflyStates {
		g.renderer.DrawFirefly(screen, state)
	}

	// 5. Dibujar punto de atracción si está activo
	if g.showAttraction {
		pulse := math.Abs(math.Sin(g.attractionPulse * math.Pi))
		g.renderer.DrawAttractionPoint(screen, g.attractionPoint, pulse)
	}

	// 6. Dibujar HUD
	fireflyCount := g.manager.GetFireflyCount()
	lanternCount := len(lanterns)
	wind := g.manager.GetWind()
	fps := g.fpsCounter.currentFPS
	isPaused := g.gameState == config.GameStatePaused

	g.uiRenderer.DrawHUD(screen, fireflyCount, lanternCount, wind, fps, isPaused)

	// 7. Dibujar controles
	g.uiRenderer.DrawControls(screen)

	// 8. Dibujar panel de objetivos
	g.uiRenderer.DrawObjectivePanel(screen, fireflyCount)

	// 9. Dibujar overlay de pausa si está pausado
	if g.gameState == config.GameStatePaused {
		g.uiRenderer.DrawPauseOverlay(screen)
	}
}

// Layout implementa ebiten.Game.Layout
// Define el tamaño lógico de la pantalla
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

// togglePause alterna entre pausado y corriendo
func (g *Game) togglePause() {
	if g.gameState == config.GameStateRunning {
		g.gameState = config.GameStatePaused
	} else if g.gameState == config.GameStatePaused {
		g.gameState = config.GameStateRunning
		g.lastUpdateTime = time.Now() // Reset delta time
	}
}

// createLantern crea un nuevo farol en la posición especificada
func (g *Game) createLantern(x, y float64) {
	success := g.manager.AddLantern(x, y)
	if !success {
		// Podríamos mostrar un mensaje de que se alcanzó el límite
	}
}

// changeWind cambia la dirección del viento
func (g *Game) changeWind() {
	cmd := manager.Command{
		Type: manager.CommandUpdateWind,
	}

	// Envío non-blocking
	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
		// Canal lleno, ignorar
	}
}

// setAttractionPoint establece un punto de atracción para las luciérnagas
func (g *Game) setAttractionPoint(x, y float64) {
	g.attractionPoint = utils.Vector2D{X: x, Y: y}
	g.showAttraction = true
	g.attractionPulse = 0.0

	cmd := manager.Command{
		Type: manager.CommandSetAttraction,
		Data: g.attractionPoint,
	}

	// Envío non-blocking
	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
		// Canal lleno, ignorar
	}
}

// clearAttractionPoint elimina el punto de atracción
func (g *Game) clearAttractionPoint() {
	g.showAttraction = false

	cmd := manager.Command{
		Type: manager.CommandClearAttraction,
	}

	// Envío non-blocking
	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
		// Canal lleno, ignorar
	}
}

// Shutdown detiene el juego y todas sus goroutines de forma limpia
func (g *Game) Shutdown() {
	g.manager.Stop()
}
