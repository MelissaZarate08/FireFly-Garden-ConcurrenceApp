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

	lastPlayerSpawn     time.Time
	playerSpawnCooldown time.Duration
}

type FPSCounter struct {
	frames       int
	lastFPSTime  time.Time
	currentFPS   float64
}

func NewFPSCounter() *FPSCounter {
	return &FPSCounter{
		frames:      0,
		lastFPSTime: time.Now(),
		currentFPS:  0,
	}
}

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

	manager.Start()

	return game
}

func (g *Game) Update() error {
	now := time.Now()
	dt := now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	g.processInput(dt)

	if g.gameState == config.GameStateRunning {
		g.updateGameLogic(dt)
	}

	g.fpsCounter.Update()

	return nil
}

func (g *Game) processInput(dt float64) {
	if g.inputHandler.IsKeyJustPressed(ebiten.KeyEscape) {
		return 
	}

	if g.inputHandler.IsKeyJustPressed(ebiten.KeyP) {
		g.togglePause()
	}

	if g.gameState == config.GameStatePaused {
		return
	}

	if g.inputHandler.IsKeyJustPressed(ebiten.KeyL) {
		mx, my := ebiten.CursorPosition()
		g.createLantern(float64(mx), float64(my))
	}

	if g.inputHandler.IsKeyJustPressed(ebiten.KeyW) {
		g.changeWind()
	}

	if g.inputHandler.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.setAttractionPoint(float64(mx), float64(my))
	}

	if g.inputHandler.IsKeyJustPressed(ebiten.KeyK) {
		if time.Since(g.lastPlayerSpawn) >= g.playerSpawnCooldown {
			mx, my := ebiten.CursorPosition()
			go g.manager.SpawnBurst(float64(mx), float64(my), config.SpawnBurstCount)
			g.lastPlayerSpawn = time.Now()
		}
	}

	if !g.inputHandler.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.showAttraction {
		g.attractionPulse += dt * 3
		if g.attractionPulse > 1.0 {
			g.clearAttractionPoint()
		}
	}
}

func (g *Game) updateGameLogic(dt float64) {
	g.manager.UpdateLanterns(dt)

	if g.showAttraction {
		g.attractionPulse += dt * 2
		if g.attractionPulse > 1.0 {
			g.attractionPulse = 0.0
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.DrawBackground(screen)

	g.renderer.DrawWind(screen, g.manager.GetWind())

	lanterns := g.manager.GetLanterns()
	for _, lantern := range lanterns {
		g.renderer.DrawLantern(screen, lantern)
	}

	fireflyStates := g.manager.GetFireflyStates()
	for _, state := range fireflyStates {
		g.renderer.DrawFirefly(screen, state)
	}

	if g.showAttraction {
		pulse := math.Abs(math.Sin(g.attractionPulse * math.Pi))
		g.renderer.DrawAttractionPoint(screen, g.attractionPoint, pulse)
	}

	fireflyCount := g.manager.GetFireflyCount()
	lanternCount := len(lanterns)
	wind := g.manager.GetWind()
	fps := g.fpsCounter.currentFPS
	isPaused := g.gameState == config.GameStatePaused

	g.uiRenderer.DrawHUD(screen, fireflyCount, lanternCount, wind, fps, isPaused)

	g.uiRenderer.DrawControls(screen)

	g.uiRenderer.DrawObjectivePanel(screen, fireflyCount)

	if g.gameState == config.GameStatePaused {
		g.uiRenderer.DrawPauseOverlay(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func (g *Game) togglePause() {
	if g.gameState == config.GameStateRunning {
		g.gameState = config.GameStatePaused
	} else if g.gameState == config.GameStatePaused {
		g.gameState = config.GameStateRunning
		g.lastUpdateTime = time.Now() 
	}
}

func (g *Game) createLantern(x, y float64) {
	success := g.manager.AddLantern(x, y)
	if !success {
	}
}

func (g *Game) changeWind() {
	cmd := manager.Command{
		Type: manager.CommandUpdateWind,
	}

	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
	}
}

func (g *Game) setAttractionPoint(x, y float64) {
	g.attractionPoint = utils.Vector2D{X: x, Y: y}
	g.showAttraction = true
	g.attractionPulse = 0.0

	cmd := manager.Command{
		Type: manager.CommandSetAttraction,
		Data: g.attractionPoint,
	}

	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
	}
}

func (g *Game) clearAttractionPoint() {
	g.showAttraction = false

	cmd := manager.Command{
		Type: manager.CommandClearAttraction,
	}

	select {
	case g.manager.GetCommandChannel() <- cmd:
	default:
	}
}

func (g *Game) Shutdown() {
	g.manager.Stop()
}
