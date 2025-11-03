package config

import "time"

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TargetFPS    = 60
)

const (
	AutoSpawnEnabled = true   
	ObjectiveCount    = 50    
)

//interaccion
const (
	SpawnBurstCount         = 6              
	PlayerSpawnCooldownSecs = 1            
)

//configuración de luciérnagas
const (
	MaxFireflies           = 100
	InitialFireflyCount    = 15                
	FireflySpawnInterval   = time.Second * 2
	FireflySize            = 8.0
	FireflySpeed           = 1.5
	FireflyBlinkCycleMin   = 1.0 
	FireflyBlinkCycleMax   = 3.0 
	FireflyAttractionForce = 0.3
	FireflyWindResistance  = 0.5

	FireflyLifespanMin = 12.0
	FireflyLifespanMax = 30.0
)

const (
	MaxLanterns           = 10
	LanternRadius         = 120.0
	LanternInfluenceForce = 0.5
	LanternSize           = 16.0
)

const (
	WindChangeInterval = time.Second * 5
	WindForce          = 0.8
	WindMaxStrength    = 2.0
)

const (
	StateChannelBuffer   = 200
	CommandChannelBuffer = 50
)

var (
	BackgroundColor  = [4]uint8{10, 15, 35, 255}      
	FireflyColorDim  = [4]uint8{180, 255, 100, 100}   
	FireflyColorFull = [4]uint8{255, 255, 150, 255}   
	LanternColor     = [4]uint8{255, 200, 100, 200}   
	WindColor        = [4]uint8{150, 150, 255, 80}    
	UITextColor      = [4]uint8{255, 255, 255, 255}  
)

//estados
const (
	GameStateRunning = iota
	GameStatePaused
	GameStateGameOver
)
