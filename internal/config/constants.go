package config

import "time"

// Configuración de pantalla
const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TargetFPS    = 60
)

// Control de comportamiento del spawner
const (
	AutoSpawnEnabled = true   // si false => solo spawner inicial
	ObjectiveCount    = 50    // objetivo jugable (meta)
)

// Jugador / interacción
const (
	SpawnBurstCount         = 6               // cuántas luciérnagas crea una ráfaga (K o al colocar farol)
	PlayerSpawnCooldownSecs = 1               // cooldown en segundos entre spawns del jugador
)

// Configuración de luciérnagas
const (
	MaxFireflies           = 100
	InitialFireflyCount    = 12                 // menos al inicio para crear desafío
	FireflySpawnInterval   = time.Second * 2
	FireflySize            = 8.0
	FireflySpeed           = 1.5
	FireflyBlinkCycleMin   = 1.0 // segundos
	FireflyBlinkCycleMax   = 3.0 // segundos
	FireflyAttractionForce = 0.3
	FireflyWindResistance  = 0.5

	// Vida de luciérnagas (en segundos)
	FireflyLifespanMin = 12.0
	FireflyLifespanMax = 30.0
)

// Configuración de faroles
const (
	MaxLanterns           = 10
	LanternRadius         = 120.0
	LanternInfluenceForce = 0.5
	LanternSize           = 16.0
)

// Configuración de viento
const (
	WindChangeInterval = time.Second * 5
	WindForce          = 0.8
	WindMaxStrength    = 2.0
)

// Configuración de canales
const (
	StateChannelBuffer   = 200
	CommandChannelBuffer = 50
)

// Colores (RGBA)
var (
	BackgroundColor  = [4]uint8{10, 15, 35, 255}      // Azul oscuro nocturno
	FireflyColorDim  = [4]uint8{180, 255, 100, 100}   // Verde-amarillo tenue
	FireflyColorFull = [4]uint8{255, 255, 150, 255}   // Amarillo brillante
	LanternColor     = [4]uint8{255, 200, 100, 200}   // Naranja cálido
	WindColor        = [4]uint8{150, 150, 255, 80}    // Azul transparente
	UITextColor      = [4]uint8{255, 255, 255, 255}   // Blanco
)

// Estados del juego
const (
	GameStateRunning = iota
	GameStatePaused
	GameStateGameOver
)
