package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/render"
)

func main() {
	// Configurar Ebiten
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("🌙 Jardín de Luciérnagas - Programación Concurrente")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(config.TargetFPS)
	
	// Crear instancia del juego
	game := render.NewGame()
	
	// Configurar manejo de señales para shutdown limpio
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Println("Señal de interrupción recibida, cerrando limpiamente...")
		game.Shutdown()
		os.Exit(0)
	}()
	
	// Mensaje de inicio
	log.Println("===========================================")
	log.Println("  🌙 JARDÍN DE LUCIÉRNAGAS")
	log.Println("  Proyecto de Programación Concurrente")
	log.Println("===========================================")
	log.Println()
	log.Println("Patrones de Concurrencia implementados:")
	log.Println("  • Fan-out/Fan-in: Luciérnagas → Agregador")
	log.Println("  • Productor-Consumidor: Manager → UI")
	log.Println("  • Worker Pool: Procesamiento paralelo")
	log.Println()
	log.Println("Mecanismos de Sincronización:")
	log.Println("  • sync.Mutex / sync.RWMutex")
	log.Println("  • sync.WaitGroup")
	log.Println("  • context.Context")
	log.Println("  • Canales buffered")
	log.Println()
	log.Println("Controles:")
	log.Println("  Click Izquierdo - Atraer luciérnagas")
	log.Println("  L - Colocar farol")
	log.Println("  W - Cambiar dirección del viento")
	log.Println("  P - Pausar/Reanudar")
	log.Println("  ESC - Salir")
	log.Println("===========================================")
	log.Println()
	log.Println("Ejecutando juego...")
	log.Println("Verifica ausencia de race conditions con: go run -race cmd/game/main.go")
	log.Println()
	
	// Ejecutar el juego
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
	
	// Shutdown limpio
	game.Shutdown()
	log.Println("Juego cerrado correctamente. ¡Adiós!")
}