# 🌙 Jardín de Luciérnagas

**Simulación interactiva de concurrencia en Go con Ebiten**  
Proyecto de Programación Concurrente

---

##  Descripción

Simulación visual e interactiva donde luciérnagas autónomas emergen, parpadean y mueren de forma natural. El jugador puede influir en su comportamiento mediante faroles y viento. Cada luciérnaga corre en su propia goroutine, demostrando patrones avanzados de concurrencia.

**Objetivo de la simulación interactiva**: Mantener 50+ luciérnagas simultáneamente usando estrategias de colocación de faroles y ráfagas.

---

## Patrones de Concurrencia

### 1. **Fan-out / Fan-in** (Principal)

**Propósito**: Distribuir trabajo en múltiples goroutines y recolectar resultados en un solo punto.

**Implementación**:
```go
// FAN-OUT: Manager crea N goroutines
func (fm *FireflyManager) spawnFirefly(x, y float64) {
    firefly := core.NewFirefly(id, x, y)
    fm.fireflies[id] = firefly
    
    // Lanzar goroutine independiente
    go firefly.Run(ctx, stateCh, lanterns, dt)
}

// FAN-IN: Agregador recibe de todas
func (sa *StateAggregator) aggregateLoop() {
    for {
        select {
        case state := <-sa.stateCh:  // Canal único
            sa.updateState(state)    // Actualiza mapa thread-safe
        }
    }
}
```

**Flujo**:
```
[Firefly 1] ──┐
[Firefly 2] ──┤
[Firefly N] ──┼──> [stateCh] ──> [StateAggregator] ──> [UI Snapshot]
```

**Archivos**: `firefly_manager.go:182`, `state_aggregator.go:42`

---

### 2. **Productor-Consumidor** (Secundario)

**Propósito**: Separar generación de datos de su procesamiento.

**Implementación**:
```go
// PRODUCTOR: Genera luciérnagas automáticamente
func (fm *FireflyManager) autoSpawner() {
    ticker := time.NewTicker(config.FireflySpawnInterval)
    for {
        select {
        case <-ticker.C:
            if fm.GetFireflyCount() < config.ObjectiveCount {
                fm.spawnFirefly(x, y)  // Produce entidades
            }
        }
    }
}

// CONSUMIDOR: UI renderiza estados
func (g *Game) Draw(screen *ebiten.Image) {
    states := g.manager.GetFireflyStates()  // Consume snapshot
    for _, state := range states {
        g.renderer.DrawFirefly(screen, state)
    }
}
```

**Archivos**: `firefly_manager.go:125`, `game.go:195`

---

### 3. **Worker Pool** (Adicional)

**Propósito**: Procesamiento paralelo de tareas con número fijo de workers.

**Implementación**:
```go
// Pool con 4 workers
pool := NewWorkerPool(4, 100, 100)

// Workers consumen del canal de trabajos
func (wp *WorkerPool) worker(workerID int) {
    for job := range wp.jobsCh {
        result := wp.processJob(job)
        wp.resultsCh <- result
    }
}
```

**Uso potencial**: Cálculos de colisión, pathfinding, procesamiento de física.

**Archivos**: `worker_pool.go:34`

---

## Mecanismos de Sincronización

### **sync.RWMutex** (3 instancias)
```go
// Protege mapa de luciérnagas
fm.firefliesMux.Lock()
fm.fireflies[id] = firefly
fm.firefliesMux.Unlock()

// Lecturas concurrentes (RLock permite múltiples lectores)
fm.firefliesMux.RLock()
count := len(fm.fireflies)
fm.firefliesMux.RUnlock()
```

**Ubicación**: `firefly_manager.go:23,26,28`

### **sync.WaitGroup**
```go
// Espera a que todas las goroutines terminen
func (fm *FireflyManager) Stop() {
    fm.cancel()          // Cancela contexto
    fm.wg.Wait()         // Espera a que finalicen
    fm.aggregator.Stop() // Limpieza
}
```

**Ubicación**: `firefly_manager.go:310`

### **context.Context**
```go
// Cancelación propagada a todas las goroutines
func (f *Firefly) Run(ctx context.Context, ...) {
    for {
        select {
        case <-ctx.Done():  // Señal de parada
            return
        case <-ticker.C:
            f.update()
        }
    }
}
```

**Ubicación**: `firefly.go:45`, `wind.go:35`, `state_aggregator.go:42`

### **Canales Buffered**
```go
// Evita bloqueos con buffer grande
stateCh := make(chan FireflyState, 200)

// Envío non-blocking
select {
case stateCh <- state:
default:
    // Canal lleno, descartar (mejor que bloquear)
}
```

**Ubicación**: `firefly.go:88`

---

## Elementos del Proyecto

### ** Luciérnagas (Firefly)**
- **Goroutine independiente** por cada entidad
- Comportamiento autónomo: movimiento errático, parpadeo sinusoidal
- **Ciclo de vida**: Nacen, envejecen (12-30s), mueren
- Publican estado cada frame al canal `stateCh`

**Código clave**:
```go
// firefly.go:45
func (f *Firefly) Run(ctx context.Context, stateCh chan<- FireflyState, ...) {
    ticker := time.NewTicker(time.Second / 60)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            f.update(lanterns, dt)
            f.age += dt
            if f.age > f.lifespan {
                return  // Muerte por edad
            }
            f.publishState(stateCh, true)
        }
    }
}
```

### ** Faroles (Lantern)**
- Puntos de atracción estáticos colocados por el jugador
- Radio de influencia: 120 píxeles
- **Generan ráfaga de 6 luciérnagas** al colocarse (feedback inmediato)

**Código clave**:
```go
// firefly_manager.go:273
func (fm *FireflyManager) AddLantern(x, y float64) bool {
    lantern := core.NewLantern(x, y)
    fm.lanterns = append(fm.lanterns, lantern)
    
    // Feedback: generar ráfaga
    go fm.SpawnBurst(x, y, 6)
    return true
}
```

### ** Viento (Wind)**
- **Goroutine independiente** que cambia dirección cada 5 segundos
- 8 direcciones cardinales (N, S, E, W, NE, NW, SE, SW)
- Afecta el movimiento de todas las luciérnagas

**Código clave**:
```go
// wind.go:35
func (w *Wind) Run(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.changeDirection()  // Cambia aleatoriamente
        }
    }
}
```

### ** Ráfagas (Burst)**
- Spawn instantáneo de múltiples luciérnagas (6 por defecto)
- **Trigger**: Colocar farol (L) o presionar K
- Cooldown: 1 segundo

**Código clave**:
```go
// firefly_manager.go:201
func (fm *FireflyManager) SpawnBurst(x, y float64, count int) {
    for i := 0; i < count; i++ {
        dx := utils.RandomFloat(-40, 40)
        dy := utils.RandomFloat(-40, 40)
        fm.spawnFirefly(x+dx, y+dy)
    }
}
```

---

## Instalación y Ejecución

### **Requisitos**
- Go 1.21+
- Sistema operativo: Windows, Linux, macOS

### **Instalación**
```bash
# Clonar repositorio
git clone <url>
cd firefly-garden

# Descargar dependencias
go mod download
```

### **Ejecución Normal**
```bash
go run cmd/game/main.go
```

### **Ejecución con Race Detector** 
```bash
go run -race cmd/game/main.go
```

### **Build para Producción**
```bash
go build -o firefly-garden cmd/game/main.go
./firefly-garden
```

---

## Controles

| Tecla/Acción | Función |
|--------------|---------|
| **Click Izquierdo** | Atraer luciérnagas al cursor |
| **L** | Colocar farol (genera ráfaga de 6) |
| **K** | Generar ráfaga cerca del cursor |
| **W** | Cambiar dirección del viento |
| **P** | Pausar/Reanudar |
| **ESC** | Salir |

---

## Métricas Mostradas en HUD

- **Luciérnagas**: Contador actual / máximo (100)
- **Faroles**: Faroles colocados / máximo (10)
- **Viento**: Dirección actual (N, S, E, W, etc.)
- **Objetivo**: Meta a alcanzar (+50)
- **FPS**: Frames por segundo
- **Goroutines**: Número de goroutines activas
- **Descartados**: Estados descartados por canal lleno (métrica de rendimiento)

---

## Estrategia Anti-Race Conditions

### **1. Snapshots Inmutables**
```go
// CORRECTO: Copia thread-safe
func (sa *StateAggregator) GetSnapshot() []FireflyState {
    sa.statesMux.RLock()
    defer sa.statesMux.RUnlock()
    
    snapshot := make([]FireflyState, 0, len(sa.states))
    for _, state := range sa.states {
        snapshot = append(snapshot, state)  // Copia, no referencia
    }
    return snapshot
}

// UI consume snapshot (nunca acceso directo)
states := g.manager.GetFireflyStates()
```

### **2. Non-blocking Channel Operations**
```go
// CORRECTO: No bloquea si canal lleno
select {
case stateCh <- state:
    // Enviado exitosamente
default:
    // Canal lleno, descartar y contabilizar
    atomic.AddUint64(&droppedStates, 1)
}
```

### **3. Context para Cancelación Limpia**
```go
// Manager cancela contexto
fm.cancel()

// Todas las goroutines reciben señal
case <-ctx.Done():
    return
```

---

## Verificación de Correctitud

### **Test de Race Conditions**
```bash
go run -race cmd/game/main.go
# Salida esperada: Sin warnings de race
```

### **Métricas de Rendimiento**
- **FPS objetivo**: 60
- **FPS real**: 60.0 (sin drops)
- **Goroutines**: ~40 (26 luciérnagas + sistema)
- **Estados descartados**: 0 (canal bien dimensionado)

---

## Configuración Avanzada

Editar `internal/config/constants.go`:

```go
// Ajustar dificultad
const ObjectiveCount = 50         // Meta del juego
const MaxFireflies = 100          // Límite de entidades
const AutoSpawnEnabled = true     // Spawner automático

// Ajustar comportamiento
const FireflyLifespanMin = 12.0   // Vida mínima (seg)
const FireflyLifespanMax = 30.0   // Vida máxima (seg)
const SpawnBurstCount = 6         // Luciérnagas por ráfaga

// Ajustar rendimiento
const StateChannelBuffer = 200    // Tamaño del canal Fan-in
const TargetFPS = 60              // FPS objetivo
```

---

## Pruebas de Estrés

### **Spawner Agresivo**
```go
// constants.go
const MaxFireflies = 200
const SpawnBurstCount = 20
```

Ejecutar con race detector y verificar:
-  Sin race conditions
-  FPS estable
-  Shutdown limpio (Ctrl+C)

---

## Características Destacadas

1. **Arquitectura limpia**: Sin funciones anónimas, código bien organizado
2. **Thread-safety completo**: Pasa `go run -race` sin errores
3. **Shutdown ordenado**: Context + WaitGroup garantizan cierre limpio
4. **Métricas avanzadas**: Contador de descartados, goroutines en tiempo real
5. **Sistema de vida**: Entidades mueren naturalmente (simulación realista)
6. **Feedback inmediato**: Ráfagas al colocar faroles

---

## Referencias Técnicas

- **Ebiten**: github.com/hajimehoshi/ebiten/v2
- **Patrones**: "Go Concurrency Patterns" (Google I/O 2012)
- **Sincronización**: Go sync package documentation

---

## Autor

**Proyecto de Programación Concurrente**  
Implementación de patrones Fan-out/Fan-in, Productor-Consumidor y Worker Pool en Go.
por: Karla Melissa Corral Zárate - Ingeniería en Software 7A

---
