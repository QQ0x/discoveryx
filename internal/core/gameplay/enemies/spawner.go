package enemies

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/constants"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/utils/math"
	"log"
	stdmath "math"
	"math/rand"
	"time"
)

// EnemyConfig contains configuration parameters for enemy spawning
type EnemyConfig struct {
	ImagePath                 string  // Path to the enemy image
	MinWallLength             float64 // Minimum wall length required for enemy placement
	MaxWallDeviation          float64 // Maximum allowed deviation in wall flatness
	MinDistanceBetweenEnemies float64 // Minimum distance between enemies
}

// Spawner handles the spawning of enemies in the game world
type Spawner struct {
	Config EnemyConfig
}

// NewSpawner creates a new enemy spawner with default configuration
func NewSpawner() *Spawner {
	return &Spawner{
		Config: EnemyConfig{
			ImagePath:                 "images/gameScene/Enemies/enemy_1.png",
			MinWallLength:             16.0, // Assuming enemy is about 16 pixels wide
			MaxWallDeviation:          10.0, // Allow 10 degree deviation in wall flatness
			MinDistanceBetweenEnemies: 32.0, // Minimum 32 pixels between enemies
		},
	}
}

// SpawnObjectsOnWalls spawns objects on walls in the visible world
func SpawnObjectsOnWalls(world *worldgen.GeneratedWorld, objectTypes []string, spawnChance float64, minDistanceBetweenObjects float64) []*Enemy {
	spawner := NewSpawner()
	// Ensure minimum distance is at least 32.0 units to prevent enemies from spawning on top of each other
	if minDistanceBetweenObjects < 32.0 {
		minDistanceBetweenObjects = 32.0
	}
	// Override the default minimum distance with the one provided
	spawner.Config.MinDistanceBetweenEnemies = minDistanceBetweenObjects
	return spawner.SpawnEnemiesOnWalls(world, objectTypes, spawnChance)
}

// SpawnEnemiesOnWalls spawns enemies on suitable walls in the visible world
func (s *Spawner) SpawnEnemiesOnWalls(world *worldgen.GeneratedWorld, enemyTypes []string, spawnChance float64) []*Enemy {
	// Random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Load enemy image to get dimensions
	enemyImage := assets.GetImage(s.Config.ImagePath)
	// Get the original dimensions
	originalWidth := float64(enemyImage.Bounds().Dx())
	originalHeight := float64(enemyImage.Bounds().Dy())
	// Apply the same 0.5 scaling factor that's used in Enemy.Draw()
	enemyWidth := originalWidth * 0.5
	enemyHeight := originalHeight * 0.5

	// List of all spawned enemies to ensure minimum distance
	spawnedPositions := []math.Vector{}
	// List of created enemy entities to return
	spawnedEnemies := []*Enemy{}

	// Get player position to determine visible chunks
	playerX, playerY := world.GetPlayerPosition()
	playerCellX := int(playerX) / worldgen.CellSize
	playerCellY := int(playerY) / worldgen.CellSize
	playerChunkX := playerCellX / worldgen.ChunkSize
	playerChunkY := playerCellY / worldgen.ChunkSize

	// Set a maximum number of enemies to spawn to prevent excessive processing
	maxTotalEnemies := 500

	// Iterate through chunks within visibility radius
	for y := playerChunkY - worldgen.VisibilityRadius; y <= playerChunkY+worldgen.VisibilityRadius; y++ {
		for x := playerChunkX - worldgen.VisibilityRadius; x <= playerChunkX+worldgen.VisibilityRadius; x++ {
			// Check if we've already spawned the maximum number of enemies
			if len(spawnedEnemies) >= maxTotalEnemies {
				return spawnedEnemies
			}

			chunk := world.GetChunk(x, y)
			if chunk == nil || !chunk.IsLoaded {
				continue
			}

			// Step 1: Collection of wall points
			if constants.DebugLogging {
				log.Printf("Step 1: Starting collection of wall points for chunk (%d, %d)", x, y)
			}

			allWallPoints := []worldgen.WallPoint{}

			// Iterate through all possible cell positions in this chunk
			for localY := 0; localY < worldgen.ChunkSize; localY++ {
				for localX := 0; localX < worldgen.ChunkSize; localX++ {
					// Calculate world coordinates for this cell
					worldX := (x*worldgen.ChunkSize + localX) * worldgen.CellSize
					worldY := (y*worldgen.ChunkSize + localY) * worldgen.CellSize

					// Get the cell at these coordinates
					cell := world.GetCellAt(worldX, worldY)
					if cell == nil {
						continue
					}

					// Get wall points in world coordinates
					wallPoints := cell.GetWallsInWorldCoordinates()
					if wallPoints == nil {
						continue
					}

					// Add all wall points to our collection
					allWallPoints = append(allWallPoints, wallPoints...)
				}
			}

			if constants.DebugLogging {
				log.Printf("Step 1: Collected %d wall points for chunk (%d, %d)", len(allWallPoints), x, y)
			}

			// Limit the number of wall points to process to prevent excessive computation
			maxWallPoints := 1000
			if len(allWallPoints) > maxWallPoints {
				// If we have too many wall points, randomly select a subset
				shuffledIndices := rng.Perm(len(allWallPoints))
				newWallPoints := make([]worldgen.WallPoint, maxWallPoints)
				for i := 0; i < maxWallPoints; i++ {
					newWallPoints[i] = allWallPoints[shuffledIndices[i]]
				}
				allWallPoints = newWallPoints
			}

			// Step 2: Connecting points to a line and segmentation
			if constants.DebugLogging {
				log.Printf("Step 2: Starting connection of wall points to segments for chunk (%d, %d)", x, y)
			}

			// Find continuous wall segments
			wallSegments := s.findWallSegments(allWallPoints)

			// Limit the number of wall segments to process
			maxSegments := 50
			if len(wallSegments) > maxSegments {
				wallSegments = wallSegments[:maxSegments]
			}

			if constants.DebugLogging {
				log.Printf("Step 2: Created %d wall segments for chunk (%d, %d)", len(wallSegments), x, y)
			}

			// For each wall segment, try to spawn enemies
			for _, segment := range wallSegments {
				// Check if the segment is long enough for an enemy
				segmentLength := s.getSegmentLength(segment)
				if constants.DebugLogging {
					log.Printf("Segment length: %.2f (minimum: %.2f)", segmentLength, s.Config.MinWallLength)
				}
				if len(segment) < 2 || segmentLength < s.Config.MinWallLength {
					if constants.DebugLogging {
						log.Printf("Segment length check failed: %.2f (minimum: %.2f)", segmentLength, s.Config.MinWallLength)
					}
					continue
				}

				// Check if the segment is flat enough
				isFlat := s.isSegmentFlat(segment, s.Config.MaxWallDeviation)
				if constants.DebugLogging {
					log.Printf("Segment flatness: %v (max deviation: %.2f)", isFlat, s.Config.MaxWallDeviation)
				}
				if !isFlat {
					if constants.DebugLogging {
						log.Printf("Segment flatness check failed (max deviation: %.2f)", s.Config.MaxWallDeviation)
					}
					continue
				}

				// Calculate how many enemies can fit on this segment
				// We need space for the enemy width plus the minimum distance between enemies
				// The formula accounts for the fact that we don't need extra space after the last enemy
				maxEnemies := int((segmentLength + s.Config.MinDistanceBetweenEnemies) / (enemyWidth + s.Config.MinDistanceBetweenEnemies))

				if constants.DebugLogging {
					log.Printf("maxEnemies calculation: maxEnemies=%d (segmentLength=%.2f, enemyWidth=%.2f, minDistance=%.2f)",
						maxEnemies, segmentLength, enemyWidth, s.Config.MinDistanceBetweenEnemies)
				}

				// If segment is too short for even one enemy, skip
				if maxEnemies < 1 {
					continue
				}

				// Limit the maximum number of enemies per segment
				if maxEnemies > 5 {
					maxEnemies = 5
				}

				// Determine how many enemies to actually spawn (based on chance)
				numToSpawn := 0
				for i := 0; i < maxEnemies; i++ {
					if rng.Float64() <= spawnChance {
						numToSpawn++
					}
				}

				if constants.DebugLogging {
					log.Printf("Spawn chance calculation: numToSpawn=%d (maxEnemies: %d, chance: %.2f)", numToSpawn, maxEnemies, spawnChance)
				}

				// If no enemies to spawn, skip
				if numToSpawn == 0 {
					if constants.DebugLogging {
						log.Printf("No enemies to spawn (numToSpawn=0), skipping segment")
					}
					continue
				}

				// Calculate spacing between enemies
				// If there's only one enemy to spawn, place it in the middle of the segment
				// Otherwise, distribute enemies evenly along the segment
				var spacing float64
				if numToSpawn == 1 {
					spacing = segmentLength / 2.0
				} else {
					// This formula ensures enemies are evenly spaced and the first/last enemies
					// are not placed exactly at the segment endpoints
					spacing = segmentLength / float64(numToSpawn+1)
				}

				// For each enemy to spawn
				for i := 1; i <= numToSpawn; i++ {
					// Check if we've already spawned the maximum number of enemies
					if len(spawnedEnemies) >= maxTotalEnemies {
						return spawnedEnemies
					}

					// Step 3: Placement of the enemy on the segment
					if constants.DebugLogging {
						log.Printf("Step 3: Starting placement of enemy %d/%d on segment (length: %.2f)", i, numToSpawn, segmentLength)
					}

					// Calculate position along the segment
					t := float64(i) * spacing / segmentLength
					spawnPos := s.interpolateSegment(segment, t)

					if constants.DebugLogging {
						log.Printf("Step 3: Calculated initial position at (%.2f, %.2f) with t=%.2f", spawnPos.X, spawnPos.Y, t)
					}

					// Check minimum distance to other spawned enemies
					tooClose := false
					for _, pos := range spawnedPositions {
						dx := spawnPos.X - pos.X
						dy := spawnPos.Y - pos.Y
						distSq := dx*dx + dy*dy

						if distSq < s.Config.MinDistanceBetweenEnemies*s.Config.MinDistanceBetweenEnemies {
							tooClose = true
							break
						}
					}

					if tooClose {
						if constants.DebugLogging {
							log.Printf("Step 3: Position too close to existing enemy, skipping")
						}
						continue
					}

					// Step 4: Calculation of rotation
					if constants.DebugLogging {
						log.Printf("Step 4: Starting calculation of rotation for enemy at (%.2f, %.2f)", spawnPos.X, spawnPos.Y)
					}

					// Calculate wall normal at this position
					normal := s.getSegmentNormalAt(segment, t)

					if constants.DebugLogging {
						log.Printf("Step 4: Got normal vector (%.2f, %.2f)", normal.X, normal.Y)
					}

					// Ensure normal is normalized
					normalX, normalY := normal.X, normal.Y
					magnitude := stdmath.Sqrt(normalX*normalX + normalY*normalY)
					if magnitude > 0 {
						normalX /= magnitude
						normalY /= magnitude
					}

					// Calculate rotation angle based on normal vector
					// Default enemy faces up (0 degrees), so we need to rotate based on normal
					// Use the normal vector directly without the extra 90-degree rotation
					// atan2 gives angle in radians, convert to degrees
					angle := stdmath.Atan2(normalX, -normalY) * 180 / stdmath.Pi

					if constants.DebugLogging {
						log.Printf("Step 4: Calculated rotation angle: %.2f degrees", angle)
					}

					// Choose random enemy type
					enemyType := enemyTypes[rng.Intn(len(enemyTypes))]

					// Initial offset spawn position in direction of normal vector
					// This is just a starting point, we'll adjust it based on transparency checks
					initialOffsetX := 5.0 // Initial offset to ensure enemy is on the wall
					spawnX := spawnPos.X + normalX*initialOffsetX
					spawnY := spawnPos.Y + normalY*initialOffsetX

					// Get the cell and snippet at the spawn position
					cell := world.GetCellAt(int(spawnX), int(spawnY))
					if cell == nil || cell.Snippet == nil {
						continue // Skip if we can't get the cell or snippet
					}

					// Step 5: Validation and adjustment of position
					if constants.DebugLogging {
						log.Printf("Step 5: Starting validation and adjustment of position for enemy at (%.2f, %.2f)", spawnX, spawnY)
					}

					// Adjust position to ensure enemy is properly anchored to the wall
					// We need to check if the bottom points of the enemy are in the rock (non-transparent)
					// and if the center of the enemy is in the air (transparent)
					validPosition := false
					maxAdjustmentAttempts := 5 // Reduced from 10 to 5 to prevent excessive iterations
					adjustmentStep := 1.0      // Step size for position adjustment

					for attempt := 0; attempt < maxAdjustmentAttempts && !validPosition; attempt++ {
						if constants.DebugLogging {
							log.Printf("Step 5: Adjustment attempt %d/%d at position (%.2f, %.2f)", attempt+1, maxAdjustmentAttempts, spawnX, spawnY)
						}

						// Calculate world coordinates relative to the cell
						relativeX := int(spawnX) % worldgen.CellSize
						relativeY := int(spawnY) % worldgen.CellSize
						if relativeX < 0 {
							relativeX += worldgen.CellSize
						}
						if relativeY < 0 {
							relativeY += worldgen.CellSize
						}

						// Calculate the bottom left and bottom right points of the enemy
						// These points should be in the rock (non-transparent)
						// First calculate the points as if the enemy is facing upward (0 degrees)
						bottomLeftX := -int(enemyWidth / 4)
						bottomLeftY := int(enemyHeight / 4)
						bottomRightX := int(enemyWidth / 4)
						bottomRightY := int(enemyHeight / 4)
						centerX := 0
						centerY := -int(enemyHeight / 4)

						// Convert angle to radians for rotation calculation
						angleRad := angle * stdmath.Pi / 180.0
						cosAngle := stdmath.Cos(angleRad)
						sinAngle := stdmath.Sin(angleRad)

						// Rotate the points according to the calculated angle
						rotatedBottomLeftX := int(float64(bottomLeftX)*cosAngle - float64(bottomLeftY)*sinAngle)
						rotatedBottomLeftY := int(float64(bottomLeftX)*sinAngle + float64(bottomLeftY)*cosAngle)
						rotatedBottomRightX := int(float64(bottomRightX)*cosAngle - float64(bottomRightY)*sinAngle)
						rotatedBottomRightY := int(float64(bottomRightX)*sinAngle + float64(bottomRightY)*cosAngle)
						rotatedCenterX := int(float64(centerX)*cosAngle - float64(centerY)*sinAngle)
						rotatedCenterY := int(float64(centerX)*sinAngle + float64(centerY)*cosAngle)

						// Add the relative position to get the final coordinates
						bottomLeftX = relativeX + rotatedBottomLeftX
						bottomLeftY = relativeY + rotatedBottomLeftY
						bottomRightX = relativeX + rotatedBottomRightX
						bottomRightY = relativeY + rotatedBottomRightY
						centerX = relativeX + rotatedCenterX
						centerY = relativeY + rotatedCenterY

						// Check if the bottom points are in the rock and the center is in the air
						bottomLeftInRock := s.isPointInRock(cell, bottomLeftX, bottomLeftY)
						bottomRightInRock := s.isPointInRock(cell, bottomRightX, bottomRightY)
						centerInAir := !s.isPointInRock(cell, centerX, centerY)

						if constants.DebugLogging {
							log.Printf("Step 5: Check points - bottomLeft(%d,%d): inRock=%v, bottomRight(%d,%d): inRock=%v, center(%d,%d): inAir=%v",
								bottomLeftX, bottomLeftY, bottomLeftInRock,
								bottomRightX, bottomRightY, bottomRightInRock,
								centerX, centerY, centerInAir)
						}

						if bottomLeftInRock && bottomRightInRock && centerInAir {
							validPosition = true
							if constants.DebugLogging {
								log.Printf("Step 5: Found valid position at (%.2f, %.2f)", spawnX, spawnY)
							}
						} else if !bottomLeftInRock || !bottomRightInRock {
							// If bottom points are not in rock, move deeper into the wall
							spawnX += normalX * adjustmentStep
							spawnY += normalY * adjustmentStep
							if constants.DebugLogging {
								log.Printf("Step 5: Bottom points not in rock, moving deeper into wall to (%.2f, %.2f)", spawnX, spawnY)
							}
						} else if !centerInAir {
							// If center is not in air, move away from the wall
							spawnX -= normalX * adjustmentStep
							spawnY -= normalY * adjustmentStep
							if constants.DebugLogging {
								log.Printf("Step 5: Center not in air, moving away from wall to (%.2f, %.2f)", spawnX, spawnY)
							}
						}
					}

					// Skip if we couldn't find a valid position
					if !validPosition {
						if constants.DebugLogging {
							log.Printf("Step 5: Could not find valid position after %d attempts, skipping", maxAdjustmentAttempts)
						}
						continue
					}

					// Check minimum distance again with the adjusted position
					tooClose = false
					for _, pos := range spawnedPositions {
						dx := spawnX - pos.X
						dy := spawnY - pos.Y
						distSq := dx*dx + dy*dy

						if distSq < s.Config.MinDistanceBetweenEnemies*s.Config.MinDistanceBetweenEnemies {
							tooClose = true
							break
						}
					}

					if tooClose {
						if constants.DebugLogging {
							log.Printf("Step 5: Adjusted position too close to existing enemy, skipping")
						}
						continue
					}

					// Step 6: Finalization of placement
					if constants.DebugLogging {
						log.Printf("Step 6: Starting finalization of enemy placement at (%.2f, %.2f) with rotation %.2f", spawnX, spawnY, angle)
					}

					// Create the enemy entity with the adjusted position
					imagePath := "images/gameScene/Enemies/" + enemyType + ".png"
					enemy := NewEnemy(enemyType, spawnX, spawnY, angle, imagePath)

					// Add to the list of spawned enemies
					spawnedEnemies = append(spawnedEnemies, enemy)
					spawnedPositions = append(spawnedPositions, math.Vector{X: spawnX, Y: spawnY})

					if constants.DebugLogging {
						log.Printf("Step 6: Successfully created enemy of type '%s' at (%.2f, %.2f) with rotation %.2f", enemyType, spawnX, spawnY, angle)
						log.Printf("Step 6: Total enemies spawned so far: %d", len(spawnedEnemies))
					}
				}
			}
		}
	}

	if constants.DebugLogging {
		log.Printf("SpawnEnemiesOnWalls: Finished spawning %d enemies", len(spawnedEnemies))
	}
	return spawnedEnemies
}

// findWallSegments groups wall points into continuous segments
func (s *Spawner) findWallSegments(wallPoints []worldgen.WallPoint) [][]worldgen.WallPoint {
	if len(wallPoints) == 0 {
		return nil
	}

	// Group wall points into segments based on proximity and normal similarity
	segments := [][]worldgen.WallPoint{}
	visited := make([]bool, len(wallPoints))

	for i := 0; i < len(wallPoints); i++ {
		if visited[i] {
			continue
		}

		// Start a new segment with this wall point
		segment := []worldgen.WallPoint{wallPoints[i]}
		visited[i] = true

		// Find all connected wall points
		s.growSegment(wallPoints, visited, &segment, i)

		// Add the segment to our list
		segments = append(segments, segment)
	}

	return segments
}

// growSegment iteratively adds connected wall points to a segment
func (s *Spawner) growSegment(wallPoints []worldgen.WallPoint, visited []bool, segment *[]worldgen.WallPoint, startIdx int) {
	// Use a queue to process points in a breadth-first manner
	queue := []int{startIdx}

	// Process points until the queue is empty
	for len(queue) > 0 {
		// Get the next point from the queue
		currentIdx := queue[0]
		queue = queue[1:] // Remove the processed point

		current := wallPoints[currentIdx]

		// Check all other wall points
		for i := 0; i < len(wallPoints); i++ {
			if visited[i] {
				continue
			}

			candidate := wallPoints[i]

			// Check if this point is close enough to be part of the same segment
			dx := candidate.X - current.X
			dy := candidate.Y - current.Y
			distSq := dx*dx + dy*dy

			// Normalize both normals before calculating dot product
			currentNormalX, currentNormalY := current.Normal.X, current.Normal.Y
			currentMagnitude := stdmath.Sqrt(currentNormalX*currentNormalX + currentNormalY*currentNormalY)
			if currentMagnitude > 0 {
				currentNormalX /= currentMagnitude
				currentNormalY /= currentMagnitude
			}

			candidateNormalX, candidateNormalY := candidate.Normal.X, candidate.Normal.Y
			candidateMagnitude := stdmath.Sqrt(candidateNormalX*candidateNormalX + candidateNormalY*candidateNormalY)
			if candidateMagnitude > 0 {
				candidateNormalX /= candidateMagnitude
				candidateNormalY /= candidateMagnitude
			}

			// Check if normals are similar enough (dot product close to 1)
			normalDot := candidateNormalX*currentNormalX + candidateNormalY*currentNormalY

			// Ensure dot product is within valid range [-1, 1]
			if normalDot < -1.0 {
				normalDot = -1.0
			} else if normalDot > 1.0 {
				normalDot = 1.0
			}

			// Convert dot product to angle in degrees
			angle := stdmath.Acos(normalDot) * 180 / stdmath.Pi

			// If point is close and has similar normal (angle less than 45 degrees), add to segment
			// 25.0 is distance squared (5 pixels)
			if distSq < 400.0 && angle < 45.0 {
				*segment = append(*segment, candidate)
				visited[i] = true
				// Add this point to the queue for processing
				queue = append(queue, i)
			}
		}
	}
}

// getSegmentLength calculates the length of a wall segment
func (s *Spawner) getSegmentLength(segment []worldgen.WallPoint) float64 {
	if len(segment) < 2 {
		return 0
	}

	length := 0.0
	for i := 1; i < len(segment); i++ {
		dx := segment[i].X - segment[i-1].X
		dy := segment[i].Y - segment[i-1].Y
		length += stdmath.Sqrt(dx*dx + dy*dy)
	}

	return length
}

// isSegmentFlat checks if a wall segment is flat enough for enemy placement
func (s *Spawner) isSegmentFlat(segment []worldgen.WallPoint, maxDeviation float64) bool {
	if len(segment) < 3 {
		return true // A segment with 2 or fewer points is always "flat"
	}

	// Calculate average normal vector
	avgNormal := math.Vector{X: 0, Y: 0}
	for _, point := range segment {
		// Ensure each normal is normalized before adding
		normalX, normalY := point.Normal.X, point.Normal.Y
		magnitude := stdmath.Sqrt(normalX*normalX + normalY*normalY)
		if magnitude > 0 {
			normalX /= magnitude
			normalY /= magnitude
		}

		avgNormal.X += normalX
		avgNormal.Y += normalY
	}
	avgNormal.X /= float64(len(segment))
	avgNormal.Y /= float64(len(segment))

	// Normalize the average
	magnitude := stdmath.Sqrt(avgNormal.X*avgNormal.X + avgNormal.Y*avgNormal.Y)
	if magnitude > 0 {
		avgNormal.X /= magnitude
		avgNormal.Y /= magnitude
	}

	// Check deviation of each point's normal from average
	for _, point := range segment {
		// Normalize point normal
		pointNormalX, pointNormalY := point.Normal.X, point.Normal.Y
		magnitude := stdmath.Sqrt(pointNormalX*pointNormalX + pointNormalY*pointNormalY)
		if magnitude > 0 {
			pointNormalX /= magnitude
			pointNormalY /= magnitude
		}

		// Calculate dot product (1 = same direction, -1 = opposite direction)
		dot := pointNormalX*avgNormal.X + pointNormalY*avgNormal.Y

		// Ensure dot product is within valid range for acos [-1, 1]
		if dot < -1.0 {
			dot = -1.0
		} else if dot > 1.0 {
			dot = 1.0
		}

		// Convert to angle in degrees
		angle := stdmath.Acos(dot) * 180 / stdmath.Pi

		if angle > maxDeviation {
			return false
		}
	}

	return true
}

// interpolateSegment finds a point at a specific position along a segment (t from 0 to 1)
func (s *Spawner) interpolateSegment(segment []worldgen.WallPoint, t float64) math.Vector {
	if len(segment) == 0 {
		return math.Vector{X: 0, Y: 0}
	}

	if len(segment) == 1 || t <= 0 {
		return math.Vector{X: segment[0].X, Y: segment[0].Y}
	}

	if t >= 1 {
		last := segment[len(segment)-1]
		return math.Vector{X: last.X, Y: last.Y}
	}

	// Calculate total segment length
	totalLength := s.getSegmentLength(segment)
	targetDist := t * totalLength

	// Walk along segment until we find the right position
	currentDist := 0.0
	for i := 1; i < len(segment); i++ {
		p1 := segment[i-1]
		p2 := segment[i]

		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		segDist := stdmath.Sqrt(dx*dx + dy*dy)

		if currentDist+segDist >= targetDist {
			// This is the segment containing our target point
			segT := (targetDist - currentDist) / segDist
			return math.Vector{
				X: p1.X + dx*segT,
				Y: p1.Y + dy*segT,
			}
		}

		currentDist += segDist
	}

	// Fallback to last point
	last := segment[len(segment)-1]
	return math.Vector{X: last.X, Y: last.Y}
}

// getSegmentNormalAt calculates the normal vector at a specific position along a segment
func (s *Spawner) getSegmentNormalAt(segment []worldgen.WallPoint, t float64) math.Vector {
	if len(segment) == 0 {
		return math.Vector{X: 0, Y: 1} // Default normal pointing up
	}

	if len(segment) == 1 || t <= 0 {
		return segment[0].Normal
	}

	if t >= 1 {
		return segment[len(segment)-1].Normal
	}

	// Calculate total segment length
	totalLength := s.getSegmentLength(segment)
	targetDist := t * totalLength

	// Walk along segment until we find the right position
	currentDist := 0.0
	for i := 1; i < len(segment); i++ {
		p1 := segment[i-1]
		p2 := segment[i]

		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		segDist := stdmath.Sqrt(dx*dx + dy*dy)

		if currentDist+segDist >= targetDist {
			// This is the segment containing our target point
			segT := (targetDist - currentDist) / segDist
			// Interpolate normal
			normal := math.Vector{
				X: p1.Normal.X + (p2.Normal.X-p1.Normal.X)*segT,
				Y: p1.Normal.Y + (p2.Normal.Y-p1.Normal.Y)*segT,
			}

			// Normalize the vector
			magnitude := stdmath.Sqrt(normal.X*normal.X + normal.Y*normal.Y)
			if magnitude > 0 {
				normal.X /= magnitude
				normal.Y /= magnitude
			}

			return normal
		}

		currentDist += segDist
	}

	// Fallback to last point's normal
	return segment[len(segment)-1].Normal
}

// isPointInRock checks if a point in a cell is in a rock (non-transparent) or in air (transparent)
func (s *Spawner) isPointInRock(cell *worldgen.WorldCell, x, y int) bool {
	if constants.DebugLogging {
		log.Printf("isPointInRock: Checking point (%d, %d)", x, y)
	}

	if cell == nil || cell.Snippet == nil || cell.Snippet.Image == nil {
		if constants.DebugLogging {
			log.Printf("isPointInRock: Invalid cell or snippet, returning false")
		}
		return false
	}

	// Get the snippet image
	img := cell.Snippet.Image

	// Get image dimensions
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	if constants.DebugLogging {
		log.Printf("isPointInRock: Image dimensions: %dx%d", width, height)
	}

	// Check if the point is within the image bounds
	if x < 0 || x >= width || y < 0 || y >= height {
		if constants.DebugLogging {
			log.Printf("isPointInRock: Point (%d, %d) is outside image bounds, returning false", x, y)
		}
		return false
	}

	// Apply rotation if needed
	if cell.Rotation != 0 {
		if constants.DebugLogging {
			log.Printf("isPointInRock: Applying rotation of %d degrees", cell.Rotation)
		}

		// Rotation angle in radians
		angle := float64(cell.Rotation) * (stdmath.Pi / 180.0)

		// Snippet center
		centerX := float64(width) / 2
		centerY := float64(height) / 2

		// Point relative to center
		relX := float64(x) - centerX
		relY := float64(y) - centerY

		// Apply inverse rotation
		cosA := stdmath.Cos(-angle)
		sinA := stdmath.Sin(-angle)
		rotX := relX*cosA - relY*sinA
		rotY := relX*sinA + relY*cosA

		// Back to absolute point
		originalX, originalY := x, y
		x = int(rotX + centerX)
		y = int(rotY + centerY)

		if constants.DebugLogging {
			log.Printf("isPointInRock: Rotated point from (%d, %d) to (%d, %d)", originalX, originalY, x, y)
		}

		// Check if the rotated point is within the image bounds
		if x < 0 || x >= width || y < 0 || y >= height {
			if constants.DebugLogging {
				log.Printf("isPointInRock: Rotated point (%d, %d) is outside image bounds, returning false", x, y)
			}
			return false
		}
	}

	// Use At() method to get the color at the specified point
	// This is more efficient than creating a new image and reading all pixel data
	_, _, _, a := img.At(x, y).RGBA()

	// If alpha > 0, the pixel is not transparent (rock)
	// The alpha value from RGBA() is in the range [0, 65535]
	isRock := a > 0

	if constants.DebugLogging {
		if isRock {
			log.Printf("isPointInRock: Point (%d, %d) is in rock (alpha=%d), returning true", x, y, a)
		} else {
			log.Printf("isPointInRock: Point (%d, %d) is in air (alpha=%d), returning false", x, y, a)
		}
	}

	return isRock
}
