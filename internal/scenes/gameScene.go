package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/constants"
	"discoveryx/internal/core/gameplay/enemies"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/core/gameplay/projectiles"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/input"
	"discoveryx/internal/rendering/shaders"
	"discoveryx/internal/utils/math"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	stdmath "math"
)

// GameScene represents the main gameplay scene with player, enemies, and world
type GameScene struct {
	player            *player.Player
	generatedWorld    *worldgen.GeneratedWorld
	cameraPosition    math.Vector
	enemies           []*enemies.Enemy
	brightnessShader  *shaders.BrightnessShader
	bullets           []*projectiles.Bullet
	timeSinceLastShot float64
	collisionManager  *physics.CollisionManager // Manages all collision detection

	// Screen shake effect for visual feedback
	shakeTimer     float64 // Time remaining for screen shake effect
	shakeAmplitude float64 // Maximum shake amplitude in pixels
	shakeFrequency float64 // Shake frequency in cycles per second
	totalTime      float64 // Total elapsed time for time-based effects
}

// NewGameScene creates a new game scene with the provided player
func NewGameScene(player *player.Player) *GameScene {
	// Create a new collision manager with a cell size of 100 units
	// This value can be tuned based on the typical size and distribution of entities
	collisionManager := physics.NewCollisionManager(100.0)

	return &GameScene{
		player:            player,
		cameraPosition:    math.Vector{X: 0, Y: 0},
		timeSinceLastShot: 0,
		collisionManager:  collisionManager,

		// Initialize screen shake effect fields
		shakeTimer:     0,
		shakeAmplitude: 0,
		shakeFrequency: 10.0, // 10 cycles per second
		totalTime:      0,
	}
}

// Initialize sets up the game scene with world generation, shaders, and enemy placement
func (s *GameScene) Initialize(state *State) error {
	generator, err := worldgen.NewWorldGenerator()
	if err != nil {
		return err
	}

	config := worldgen.DefaultWorldGenConfig()

	s.generatedWorld, err = worldgen.NewGeneratedWorld(
		state.World.GetWidth(),
		state.World.GetHeight(),
		generator,
		config,
	)
	if err != nil {
		return err
	}

	s.brightnessShader, err = shaders.NewBrightnessShader()
	if err != nil {
		return err
	}

	objectTypes := []string{"enemy_1"}
	s.enemies = enemies.SpawnObjectsOnWalls(s.generatedWorld, objectTypes, 1.0, 32.0)

	// Position the player on the main path first
	if len(s.generatedWorld.GetWorldMap().MainPathCells) > 0 {
		// Get a position from the middle of the main path
		mainPathCells := s.generatedWorld.GetWorldMap().MainPathCells
		startIdx := len(mainPathCells) / 2
		mainPathCell := mainPathCells[startIdx]
		mainPathPos := math.Vector{
			X: float64(mainPathCell.X*worldgen.CellSize + worldgen.CellSize/2),
			Y: float64(mainPathCell.Y*worldgen.CellSize + worldgen.CellSize/2),
		}

		// Set the player's initial position
		s.player.SetPosition(mainPathPos)
	}

	// Register walls with the collision manager after positioning the player
	s.registerWalls()

	// Now try to find a better position for the player if needed
	if len(s.generatedWorld.GetWorldMap().MainPathCells) > 0 {
		// Try to find a valid spawn position on the main path
		validPosition := false
		mainPathCells := s.generatedWorld.GetWorldMap().MainPathCells

		// Start from the middle of the path and try positions in both directions
		startIdx := len(mainPathCells) / 2
		offset := 0
		maxTries := len(mainPathCells)

		for tries := 0; tries < maxTries && !validPosition; tries++ {
			// Alternate between moving forward and backward from the middle
			idx := startIdx
			if tries%2 == 0 {
				idx = startIdx + offset
				if idx >= len(mainPathCells) {
					idx = len(mainPathCells) - 1
				}
			} else {
				idx = startIdx - offset
				if idx < 0 {
					idx = 0
				}
				offset++
			}

			mainPathCell := mainPathCells[idx]
			mainPathPos := math.Vector{
				X: float64(mainPathCell.X*worldgen.CellSize + worldgen.CellSize/2),
				Y: float64(mainPathCell.Y*worldgen.CellSize + worldgen.CellSize/2),
			}

			// Check if this position collides with any walls using world data
			collision, _, _ := s.CheckWallCollisionUsingWorldData(s.player, mainPathPos)

			if !collision {
				// Found a valid position
				s.player.SetPosition(mainPathPos)
				validPosition = true
			}
		}

		// If we couldn't find a valid position, we'll keep the initial position set earlier
	}

	// Register the player with the collision manager
	s.collisionManager.RegisterEntity(s.player, s.player.GetCollider())

	// Register enemies with the collision manager
	for _, enemy := range s.enemies {
		s.collisionManager.RegisterEntity(enemy, enemy.GetCollider())
	}

	return nil
}

// registerWalls extracts wall points from the generated world, converts them to wall colliders,
// and registers them with the collision manager.
func (s *GameScene) registerWalls() {
	// Clear existing walls
	s.collisionManager.ClearWalls()

	// Create a wall collider generator with a minimum wall size of 10 units
	wallGenerator := physics.NewWallColliderGenerator(10.0)

	// Get player position in cell coordinates
	playerPos := s.player.GetPosition()
	playerCellX := int(playerPos.X) / worldgen.CellSize
	playerCellY := int(playerPos.Y) / worldgen.CellSize

	// Convert player cell coordinates to chunk coordinates
	playerChunkX := playerCellX / worldgen.ChunkSize
	playerChunkY := playerCellY / worldgen.ChunkSize

	// Use the same visibility radius as the world generator
	chunkRadius := worldgen.VisibilityRadius

	// Process all chunks in the visibility radius
	for chunkY := playerChunkY - chunkRadius; chunkY <= playerChunkY+chunkRadius; chunkY++ {
		for chunkX := playerChunkX - chunkRadius; chunkX <= playerChunkX+chunkRadius; chunkX++ {
			// Get the chunk at these coordinates
			chunk := s.generatedWorld.GetChunk(chunkX, chunkY)
			if chunk == nil || !chunk.IsLoaded {
				continue
			}

			// Process all cells in this chunk
			for _, cell := range chunk.Cells {
				if cell == nil || cell.Snippet == nil {
					continue
				}

				// Get wall points in world coordinates
				wallPoints := cell.GetWallsInWorldCoordinates()
				if len(wallPoints) == 0 {
					continue
				}

				// Convert wall points to physics.WallPoint
				physicsWallPoints := make([]physics.WallPoint, len(wallPoints))
				for i, wp := range wallPoints {
					physicsWallPoints[i] = physics.WallPoint{
						X:      wp.X,
						Y:      wp.Y,
						Normal: wp.Normal,
					}
				}

				// Generate wall colliders
				wallColliders := wallGenerator.GenerateWallColliders(physicsWallPoints, float64(worldgen.CellSize))

				// Register wall colliders with the collision manager
				for _, collider := range wallColliders {
					s.collisionManager.RegisterWall(collider)
				}
			}
		}
	}

	// Optimize walls to reduce the number of colliders
	s.collisionManager.OptimizeWalls()
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// CheckWallCollisionUsingWorldData performs collision detection between an entity and walls using world generation data.
// This is more efficient than using physics queries as it directly uses the tile information.
func (s *GameScene) CheckWallCollisionUsingWorldData(entity interface{}, plannedPos math.Vector) (bool, math.Vector, bool) {
	// Check if the entity implements GetAABBCollider
	type aabbGetter interface {
		GetAABBCollider() physics.AABBCollider
	}

	aabbEntity, ok := entity.(aabbGetter)
	if !ok {
		return false, math.Vector{}, false
	}

	// Get the entity's AABB collider
	entityCollider := aabbEntity.GetAABBCollider()

	// Calculate the entity's bounds
	minX := plannedPos.X - entityCollider.Width/2
	maxX := plannedPos.X + entityCollider.Width/2
	minY := plannedPos.Y - entityCollider.Height/2
	maxY := plannedPos.Y + entityCollider.Height/2

	// Get the cells that the entity overlaps with
	cellMinX := int(minX) / worldgen.CellSize
	cellMaxX := int(maxX) / worldgen.CellSize
	cellMinY := int(minY) / worldgen.CellSize
	cellMaxY := int(maxY) / worldgen.CellSize

	// Check each cell for wall collisions
	collision := false
	smallestDepth := 1000000.0
	var separationVector math.Vector
	var isXAxis bool

	for cellY := cellMinY; cellY <= cellMaxY; cellY++ {
		for cellX := cellMinX; cellX <= cellMaxX; cellX++ {
			// Get the cell at these coordinates
			cell := s.generatedWorld.GetCellAt(cellX*worldgen.CellSize, cellY*worldgen.CellSize)
			if cell == nil || cell.Snippet == nil {
				continue
			}

			// Get wall points in world coordinates
			wallPoints := cell.GetWallsInWorldCoordinates()
			if len(wallPoints) == 0 {
				continue
			}

			// Check for collisions with each wall point
			for _, wallPoint := range wallPoints {
				// Calculate the distance from the wall point to the entity's AABB
				// We need to check if the wall point is inside or close to the entity's AABB

				// Calculate the closest point on the AABB to the wall point
				// Clamp X value between minX and maxX
				closestX := wallPoint.X
				if closestX < minX {
					closestX = minX
				} else if closestX > maxX {
					closestX = maxX
				}

				// Clamp Y value between minY and maxY
				closestY := wallPoint.Y
				if closestY < minY {
					closestY = minY
				} else if closestY > maxY {
					closestY = maxY
				}

				// Calculate the distance from the wall point to the closest point on the AABB
				distX := wallPoint.X - closestX
				distY := wallPoint.Y - closestY

				// If the wall point is inside or very close to the AABB, we have a collision
				// We use a small threshold to account for floating point errors
				threshold := 1.0 // 1 pixel threshold
				if (distX*distX + distY*distY) <= threshold*threshold {
					collision = true

					// Calculate the penetration depth
					depth := 0.0
					normal := wallPoint.Normal
					isCollisionXAxis := false

					// Calculate penetration depth based on the normal direction
					if abs(normal.X) > abs(normal.Y) {
						// X-axis collision
						if normal.X > 0 {
							depth = maxX - wallPoint.X
						} else {
							depth = wallPoint.X - minX
						}
						isCollisionXAxis = true
					} else {
						// Y-axis collision
						if normal.Y > 0 {
							depth = maxY - wallPoint.Y
						} else {
							depth = wallPoint.Y - minY
						}
						isCollisionXAxis = false
					}

					// If this is the smallest penetration depth so far, update the separation vector
					if depth < smallestDepth {
						smallestDepth = depth
						separationVector = math.Vector{
							X: normal.X * depth,
							Y: normal.Y * depth,
						}
						isXAxis = isCollisionXAxis
					}
				}
			}
		}
	}

	return collision, separationVector, isXAxis
}

// Update handles the game logic and camera movement for the scene
func (s *GameScene) Update(state *State) error {
	// Update total elapsed time
	s.totalTime += state.DeltaTime

	if s.generatedWorld == nil {
		if err := s.Initialize(state); err != nil {
			return err
		}
	}

	// Check if player's health is zero and redirect to start screen
	if s.player.GetHealth() <= 0 {
		state.SceneManager.GoToScene(NewStartScene())
		return nil
	}

	// Get the player's current position before updating
	currentPosition := s.player.GetPosition()

	// Update player's state (input, rotation, etc.) but don't apply movement yet
	if err := s.player.Update(state.Input, state.DeltaTime); err != nil {
		return err
	}

	// Get the player's updated position after input processing
	updatedPosition := s.player.GetPosition()

	// Check for wall collisions using AABB collision detection
	// First check X-axis movement
	plannedXPosition := math.Vector{
		X: updatedPosition.X,
		Y: currentPosition.Y,
	}

	collisionX, separationVectorX, _ := s.CheckWallCollisionUsingWorldData(s.player, plannedXPosition)

	// Apply X-axis movement if no collision
	if collisionX {
		// Collision on X-axis, keep the player at the wall edge
		s.player.SetPosition(math.Vector{
			X: plannedXPosition.X + separationVectorX.X,
			Y: currentPosition.Y,
		})
	} else {
		// No collision on X-axis, allow movement
		s.player.SetPosition(math.Vector{
			X: updatedPosition.X,
			Y: currentPosition.Y,
		})
	}

	// Now check Y-axis movement
	currentXPosition := s.player.GetPosition().X
	plannedYPosition := math.Vector{
		X: currentXPosition,
		Y: updatedPosition.Y,
	}

	collisionY, separationVectorY, _ := s.CheckWallCollisionUsingWorldData(s.player, plannedYPosition)

	// Apply Y-axis movement if no collision
	if collisionY {
		// Collision on Y-axis, keep the player at the wall edge
		s.player.SetPosition(math.Vector{
			X: currentXPosition,
			Y: plannedYPosition.Y + separationVectorY.Y,
		})
	} else {
		// No collision on Y-axis, allow movement
		s.player.SetPosition(math.Vector{
			X: currentXPosition,
			Y: updatedPosition.Y,
		})
	}

	// Final check to ensure player is not inside a wall
	// This handles cases where minimal overlaps might still occur due to player hitbox size
	// We limit the number of iterations to prevent infinite loops when player is stuck between two walls
	const maxIterations = 3 // Maximum number of collision resolution attempts
	var iterations int = 0
	var finalPosition math.Vector
	var finalCollision bool
	var separationVector math.Vector

	// Try to resolve collisions up to maxIterations times
	for iterations < maxIterations {
		finalPosition = s.player.GetPosition()
		finalCollision, separationVector, _ = s.CheckWallCollisionUsingWorldData(s.player, finalPosition)

		if !finalCollision {
			// No collision, we're done
			break
		}

		// Calculate the normal vector (normalized direction of the separation vector)
		magnitude := stdmath.Sqrt(separationVector.X*separationVector.X + separationVector.Y*separationVector.Y)
		if magnitude <= 0 {
			// Invalid magnitude, can't resolve
			break
		}

		normalVector := math.Vector{
			X: separationVector.X / magnitude,
			Y: separationVector.Y / magnitude,
		}

		// Player is still inside a wall, push them out using ResolveCollision
		// Use the magnitude of the separation vector as the depth
		correctedPosition := physics.ResolveCollision(finalPosition, normalVector, magnitude)
		s.player.SetPosition(correctedPosition)

		// Debug output to help diagnose collision issues
		if constants.DebugPlayerWallCollision {
			fmt.Printf("Collision resolved (iteration %d): Player moved from (%.4f, %.4f) to (%.4f, %.4f)\n", 
				iterations+1, finalPosition.X, finalPosition.Y, correctedPosition.X, correctedPosition.Y)
		}

		iterations++
	}

	// If we hit the iteration limit, log it for debugging
	if iterations == maxIterations && constants.DebugPlayerWallCollision {
		fmt.Printf("WARNING: Hit maximum collision resolution iterations (%d). Player may be stuck.\n", maxIterations)
	}

	// If there was a collision, reduce the player's velocity
	if collisionX || collisionY || finalCollision {
		// Reduce velocity to simulate friction with the wall
		currentVelocity := s.player.GetVelocity()
		s.player.SetVelocity(currentVelocity * 0) // Set to zero to prevent further movement into walls

		// Debug output for collision velocity adjustment
		if constants.DebugPlayerWallCollision {
			fmt.Printf("Wall collision detected: Player velocity reduced from %.2f to 0\n", currentVelocity)
		}
	}

	// Update the player's collider in the collision manager
	s.collisionManager.UpdateEntity(s.player, s.player.GetCollider())

	// Update and check enemies
	var activeEnemies []*enemies.Enemy
	for _, enemy := range s.enemies {
		if enemy.Update(state.DeltaTime) {
			// Enemy should be removed (death animation completed)
			s.collisionManager.RemoveEntity(enemy)
			continue
		}

		// Update enemy's collider in the collision manager
		s.collisionManager.UpdateEntity(enemy, enemy.GetCollider())

		// Check for collision between player and enemy using the collision manager
		if !s.player.IsInvincible() {
			// Use the collision manager to check for collision
			// We use the player's collider radius plus a small buffer to ensure accurate detection
			collision, collidedEntity := s.collisionManager.CheckCollision(enemy, enemy.GetCollider().Radius+5.0)
			if collision && collidedEntity == s.player {
				// Player hit by enemy, apply damage
				s.player.TakeDamage(player.EnemyCollisionDamage)
			}
		}

		activeEnemies = append(activeEnemies, enemy)
	}
	s.enemies = activeEnemies

	// Handle player shooting and enemy shooting
	s.handleShooting(state)
	s.handleEnemyShooting(state)

	// Update and check bullets
	var activeBullets []*projectiles.Bullet
	for _, b := range s.bullets {
		// Update bullet position and check if it's still active
		shouldRemove := b.Update(state.DeltaTime)

		if shouldRemove {
			continue
		}

		// Get bullet collider
		bulletCollider := b.GetCollider()

		// Store the bullet's previous position for continuous collision detection
		prevPosition := math.Vector{
			X: bulletCollider.Position.X - stdmath.Sin(b.Rotation)*b.GetSpeed()*state.DeltaTime*60.0,
			Y: bulletCollider.Position.Y - stdmath.Cos(b.Rotation)*-b.GetSpeed()*state.DeltaTime*60.0,
		}

		if b.IsPlayerBullet {
			// Check for collision with enemies (only for player bullets)
			bulletHit := false

			// Use the collision manager to find nearby enemies
			nearbyEntities := s.collisionManager.GetNearbyEntities(bulletCollider.Position, bulletCollider.Radius+40.0)
			for _, entity := range nearbyEntities {
				enemy, isEnemy := entity.(*enemies.Enemy)
				if !isEnemy {
					continue
				}

				// Get the enemy's collider
				enemyCollider := enemy.GetCollider()

				// Use continuous collision detection to check for collision with this enemy
				// For simplicity, we assume the enemy is stationary during this frame
				collision, _, _, _ := physics.CheckContinuousCircleCircleCollision(
					prevPosition, bulletCollider.Position, bulletCollider.Radius,
					enemyCollider.Position, enemyCollider.Position, enemyCollider.Radius)

				if collision {
					// Enemy hit by player bullet, apply damage
					if enemy.TakeDamage(b.Damage) {
						// Enemy died from this damage
						// Score or other game effects could be added here
					}
					bulletHit = true
					break
				}
			}

			if bulletHit {
				// Don't add this bullet to active bullets (it hit an enemy)
				continue
			}
		} else {
			// Check for collision with player (only for enemy bullets)
			if !s.player.IsInvincible() {
				// Get the player's collider
				playerCollider := s.player.GetCollider()

				// Use continuous collision detection to check for collision with the player
				// For simplicity, we assume the player's position doesn't change significantly during this frame
				collision, _, _, _ := physics.CheckContinuousCircleCircleCollision(
					prevPosition, bulletCollider.Position, bulletCollider.Radius,
					playerCollider.Position, playerCollider.Position, playerCollider.Radius)

				if collision {
					// Player hit by enemy bullet, apply damage
					s.player.TakeDamage(b.Damage)
					// Don't add this bullet to active bullets (it hit the player)
					continue
				}
			}
		}

		// Check for collision with walls using continuous collision detection
		// We reuse the previous position calculated earlier

		// Get nearby walls
		wallColliders := s.collisionManager.GetNearbyWalls(bulletCollider.Position, bulletCollider.Radius+40.0)

		// Check for collision with any wall
		bulletHitWall := false

		for _, wall := range wallColliders {
			// Use continuous collision detection to check for collision with this wall
			collision, _, _, _ := physics.CheckContinuousCircleCollision(
				prevPosition, bulletCollider.Position, bulletCollider.Radius, wall)

			if collision {
				bulletHitWall = true
				break
			}
		}

		if bulletHitWall {
			// Don't add this bullet to active bullets (it hit a wall)
			// In a more advanced implementation, we could add visual effects at the collision point
			// or bounce the bullet off the wall using the collision normal
			continue
		}

		// Keep the bullet if it shouldn't be removed
		activeBullets = append(activeBullets, b)
	}
	s.bullets = activeBullets

	// The AABB collision detection is now handled earlier in the Update method
	// We've replaced the continuous collision detection with axis-separated AABB collision detection

	position := s.player.GetPosition()
	screenWidth := float64(state.World.GetWidth())
	screenHeight := float64(state.World.GetHeight())

	s.generatedWorld.SetPlayerPosition(position.X, position.Y)

	// Register walls from newly loaded chunks with the collision manager
	s.registerWalls()

	// Camera system implementation
	playerVelocity := s.player.GetVelocity()
	cameraTargetX := -s.cameraPosition.X
	cameraTargetY := -s.cameraPosition.Y
	offsetX := position.X - cameraTargetX
	offsetY := position.Y - cameraTargetY
	deadZoneWidth := screenWidth * constants.CameraDeadZoneX
	deadZoneHeight := screenHeight * constants.CameraDeadZoneY

	var newCameraTargetX, newCameraTargetY float64

	if playerVelocity >= constants.CameraVelocityThreshold {
		// Use deadzone-based camera following for normal movement speed
		newCameraTargetX = cameraTargetX
		newCameraTargetY = cameraTargetY

		if stdmath.Abs(offsetX) > deadZoneWidth/2 {
			if offsetX > 0 {
				newCameraTargetX = position.X - deadZoneWidth/2
			} else {
				newCameraTargetX = position.X + deadZoneWidth/2
			}
		}

		if stdmath.Abs(offsetY) > deadZoneHeight/2 {
			if offsetY > 0 {
				newCameraTargetY = position.Y - deadZoneHeight/2
			} else {
				newCameraTargetY = position.Y + deadZoneHeight/2
			}
		}
	} else {
		// Gradually center on player when moving slowly
		centeringFactor := (constants.CameraVelocityThreshold - playerVelocity) / constants.CameraVelocityThreshold
		newCameraTargetX = cameraTargetX + (position.X-cameraTargetX)*centeringFactor*constants.CameraCenteringStrength
		newCameraTargetY = cameraTargetY + (position.Y-cameraTargetY)*centeringFactor*constants.CameraCenteringStrength
	}

	targetCameraX := -newCameraTargetX
	targetCameraY := -newCameraTargetY

	// Frame-rate independent camera smoothing
	interpolationFactor := 1.0 - stdmath.Pow(1.0-constants.CameraInterpolationFactor, state.DeltaTime*60.0)
	s.cameraPosition.X += (targetCameraX - s.cameraPosition.X) * interpolationFactor
	s.cameraPosition.Y += (targetCameraY - s.cameraPosition.Y) * interpolationFactor

	// Update screen shake effect
	if s.shakeTimer > 0 {
		// Decrease shake timer
		s.shakeTimer -= state.DeltaTime
		if s.shakeTimer < 0 {
			s.shakeTimer = 0
			s.shakeAmplitude = 0
		} else {
			// Calculate shake offset based on sine wave
			// Use time and frequency to determine the phase of the sine wave
			shakePhaseX := s.shakeFrequency * s.totalTime * 2.0 * stdmath.Pi
			shakePhaseY := s.shakeFrequency*s.totalTime*2.0*stdmath.Pi + stdmath.Pi/2.0 // 90 degrees offset for Y

			// Calculate shake intensity (decreases as timer approaches 0)
			shakeIntensity := s.shakeAmplitude * (s.shakeTimer / 0.3) // 0.3 is the initial shake duration

			// Apply shake offset to camera position
			s.cameraPosition.X += stdmath.Sin(shakePhaseX) * shakeIntensity
			s.cameraPosition.Y += stdmath.Sin(shakePhaseY) * shakeIntensity
		}
	}

	return nil
}

// Draw renders the game scene with background, world, entities and lighting effects
func (s *GameScene) Draw(screen *ebiten.Image, state *State) {
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()
	tempScreen := ebiten.NewImage(worldWidth, worldHeight)

	// Scale background to fit screen while maintaining aspect ratio
	bgOp := &ebiten.DrawImageOptions{}
	gameBg := assets.GetGameBackground()
	bgWidth := float64(gameBg.Bounds().Dx())
	bgHeight := float64(gameBg.Bounds().Dy())

	scaleX := float64(worldWidth) / bgWidth
	scaleY := float64(worldHeight) / bgHeight
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	bgOp.GeoM.Scale(scale, scale)

	scaledWidth := bgWidth * scale
	scaledHeight := bgHeight * scale
	bgOp.GeoM.Translate((float64(worldWidth)-scaledWidth)/2, (float64(worldHeight)-scaledHeight)/2)

	tempScreen.DrawImage(gameBg, bgOp)

	if s.generatedWorld != nil {
		s.generatedWorld.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)
	}

	for _, enemy := range s.enemies {
		enemy.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y, worldWidth, worldHeight)
	}

	for _, b := range s.bullets {
		b.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y, worldWidth, worldHeight)
	}

	s.player.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)

	// Apply lighting effect with brightness shader
	if s.brightnessShader != nil {
		playerPos := s.player.GetPosition()
		screenPosX := playerPos.X + s.cameraPosition.X + float64(worldWidth)/2
		screenPosY := playerPos.Y + s.cameraPosition.Y + float64(worldHeight)/2

		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = tempScreen
		op.Uniforms = map[string]any{
			"PlayerPos": []float32{float32(screenPosX), float32(screenPosY)},
			"Radius":    float32(float64(worldWidth) * 0.75),
		}

		screen.DrawRectShader(worldWidth, worldHeight, s.brightnessShader.Shader(), op)
	} else {
		screen.DrawImage(tempScreen, nil)
	}

	// Draw health bar with margins from the edges
	healthBarHeight := 5.0 // Very thin bar

	// Add margins from the edges (10% of screen width for left/right, 10px from top)
	marginX := float64(worldWidth) * 0.1
	marginY := 10.0

	// Calculate health bar width with margins
	healthBarWidth := float64(worldWidth) - (marginX * 2)

	// Calculate the width of the green part based on player's health percentage
	healthPercentage := s.player.GetHealth() / player.MaxPlayerHealth
	greenWidth := healthBarWidth * healthPercentage

	// Draw the red background (lost health)
	redBar := ebiten.NewImage(int(healthBarWidth), int(healthBarHeight))
	redBar.Fill(color.RGBA{255, 0, 0, 255}) // Red color

	// Draw the red bar first (full width)
	redOp := &ebiten.DrawImageOptions{}
	redOp.GeoM.Translate(marginX, marginY) // Apply margins
	screen.DrawImage(redBar, redOp)

	// Only draw the green health bar if the player has health
	if healthPercentage > 0 {
		// Ensure the green bar has at least 1 pixel width
		if greenWidth < 1 {
			greenWidth = 1
		}

		// Draw the green foreground (remaining health)
		greenBar := ebiten.NewImage(int(greenWidth), int(healthBarHeight))
		greenBar.Fill(color.RGBA{0, 255, 0, 255}) // Green color

		// Then draw the green bar on top (partial width based on health)
		greenOp := &ebiten.DrawImageOptions{}
		greenOp.GeoM.Translate(marginX, marginY) // Apply margins
		screen.DrawImage(greenBar, greenOp)
	}
}

const fireInterval = 0.25
const enemyFireInterval = 0.5
const enemyShootRadius = 150.0

// handleShooting manages bullet firing based on keyboard or touch input
func (s *GameScene) handleShooting(state *State) {
	touch := state.Input.Touch()

	if inpututil.IsKeyJustPressed(input.KeySpace) || (touch != nil && touch.IsFireJustSwiped()) {
		s.spawnBullet()
		s.timeSinceLastShot = 0
		return
	}

	holding := state.Input.Keyboard().IsKeyPressed(input.KeySpace)
	if touch != nil && touch.IsFireHolding() {
		holding = true
	}

	if holding {
		s.timeSinceLastShot += state.DeltaTime
		if s.timeSinceLastShot >= fireInterval {
			s.spawnBullet()
			s.timeSinceLastShot = 0
		}
	} else {
		s.timeSinceLastShot = fireInterval
	}
}

// handleEnemyShooting makes enemies fire bullets at the player when in range
func (s *GameScene) handleEnemyShooting(state *State) {
	playerPos := s.player.GetPosition()
	for _, enemy := range s.enemies {
		dx := playerPos.X - enemy.Position.X
		dy := playerPos.Y - enemy.Position.Y
		if dx*dx+dy*dy <= enemyShootRadius*enemyShootRadius {
			enemy.TimeSinceLastShot += state.DeltaTime
			if enemy.TimeSinceLastShot >= enemyFireInterval {
				rot := stdmath.Atan2(-dy, dx)
				// Create a copy of the enemy position to ensure bullets spawn from the enemy
				// Add an offset in the direction of the player to make bullets spawn from the edge of the enemy
				offsetDistance := 10.0 // Distance from enemy center to spawn the bullet
				offsetX := stdmath.Sin(rot) * offsetDistance
				offsetY := stdmath.Cos(rot) * -offsetDistance
				enemyPos := math.Vector{X: enemy.Position.X + offsetX, Y: enemy.Position.Y + offsetY}
				s.bullets = append(s.bullets, projectiles.NewLinearBullet(enemyPos, rot, assets.EnemyBullet, false))
				enemy.TimeSinceLastShot = 0
			}
		} else if enemy.TimeSinceLastShot > enemyFireInterval {
			enemy.TimeSinceLastShot = enemyFireInterval
		}
	}
}

// spawnBullet creates a new bullet at the player's position with the player's rotation
func (s *GameScene) spawnBullet() {
	pos := s.player.GetPosition()
	rot := s.player.GetRotation()

	// Add an offset in the direction the player is facing to make bullets spawn from the edge of the player
	offsetDistance := 20.0 // Distance from player center to spawn the bullet
	offsetX := stdmath.Sin(rot) * offsetDistance
	offsetY := stdmath.Cos(rot) * -offsetDistance
	bulletPos := math.Vector{X: pos.X + offsetX, Y: pos.Y + offsetY}

	// Create a new bullet
	bullet := projectiles.NewBullet(bulletPos, rot, assets.PlayerBullet, true)

	// Add the bullet to the game
	s.bullets = append(s.bullets, bullet)
}
