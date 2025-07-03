package enemies

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/utils/math"
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
	enemyWidth := float64(enemyImage.Bounds().Dx())
	enemyHeight := float64(enemyImage.Bounds().Dy())

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

	// Iterate through chunks within visibility radius
	for y := playerChunkY - worldgen.VisibilityRadius; y <= playerChunkY+worldgen.VisibilityRadius; y++ {
		for x := playerChunkX - worldgen.VisibilityRadius; x <= playerChunkX+worldgen.VisibilityRadius; x++ {
			chunk := world.GetChunk(x, y)
			if chunk == nil || !chunk.IsLoaded {
				continue
			}

			// Collect all wall points from this chunk
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

			// Find continuous wall segments
			wallSegments := s.findWallSegments(allWallPoints)

			// For each wall segment, try to spawn enemies
			for _, segment := range wallSegments {
				// Check if the segment is long enough for an enemy
				if len(segment) < 2 || s.getSegmentLength(segment) < s.Config.MinWallLength {
					continue
				}

				// Check if the segment is flat enough
				if !s.isSegmentFlat(segment, s.Config.MaxWallDeviation) {
					continue
				}

				// Calculate how many enemies can fit on this segment
				segmentLength := s.getSegmentLength(segment)
				// We need space for the enemy width plus the minimum distance between enemies
				// The formula accounts for the fact that we don't need extra space after the last enemy
				maxEnemies := int((segmentLength + s.Config.MinDistanceBetweenEnemies) / (enemyWidth + s.Config.MinDistanceBetweenEnemies))

				// If segment is too short for even one enemy, skip
				if maxEnemies < 1 {
					continue
				}

				// Determine how many enemies to actually spawn (based on chance)
				numToSpawn := 0
				for i := 0; i < maxEnemies; i++ {
					if rng.Float64() <= spawnChance {
						numToSpawn++
					}
				}

				// If no enemies to spawn, skip
				if numToSpawn == 0 {
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
					// Calculate position along the segment
					t := float64(i) * spacing / segmentLength
					spawnPos := s.interpolateSegment(segment, t)

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
						continue
					}

					// Calculate wall normal at this position
					normal := s.getSegmentNormalAt(segment, t)

					// Ensure normal is normalized
					normalX, normalY := normal.X, normal.Y
					magnitude := stdmath.Sqrt(normalX*normalX + normalY*normalY)
					if magnitude > 0 {
						normalX /= magnitude
						normalY /= magnitude
					}

					// Calculate rotation angle based on normal vector
					// Default enemy faces up (0 degrees), so we need to rotate based on normal
					// To align with the wall surface (not perpendicular to it), we need to rotate by 90 degrees
					// Calculate tangent vector (parallel to wall surface) by rotating normal by 90 degrees
					tangentX := -normalY
					tangentY := normalX
					// atan2 gives angle in radians, convert to degrees
					angle := stdmath.Atan2(tangentX, -tangentY) * 180 / stdmath.Pi

					// Choose random enemy type
					enemyType := enemyTypes[rng.Intn(len(enemyTypes))]

					// Offset based on the deepest point of the segment
					deepest := stdmath.Inf(1)
					for _, p := range segment {
						d := p.X*normalX + p.Y*normalY
						if d < deepest {
							deepest = d
						}
					}
					currentD := spawnPos.X*normalX + spawnPos.Y*normalY
					offset := deepest - enemyHeight/2 - currentD
					spawnX := spawnPos.X + normalX*offset
					spawnY := spawnPos.Y + normalY*offset

					// Validate spawn position: ensure bottom corners are in rock and center is transparent
					tangent := math.Vector{X: -normalY, Y: normalX}
					valid := false
					attempts := 0
					for attempts < 5 {
						blX := spawnX - tangent.X*enemyWidth/2 - normalX*enemyHeight/2
						blY := spawnY - tangent.Y*enemyWidth/2 - normalY*enemyHeight/2
						brX := spawnX + tangent.X*enemyWidth/2 - normalX*enemyHeight/2
						brY := spawnY + tangent.Y*enemyWidth/2 - normalY*enemyHeight/2
						if isSolid(world, blX, blY) && isSolid(world, brX, brY) && isTransparent(world, spawnX, spawnY) {
							valid = true
							break
						}
						if !isSolid(world, blX, blY) || !isSolid(world, brX, brY) {
							spawnX -= normalX
							spawnY -= normalY
						} else if !isTransparent(world, spawnX, spawnY) {
							spawnX += normalX
							spawnY += normalY
						}
						attempts++
					}
					if !valid {
						continue
					}

					// Create the enemy entity
					imagePath := "images/gameScene/Enemies/" + enemyType + ".png"
					enemy := NewEnemy(enemyType, spawnX, spawnY, angle, imagePath)

					// Add to the list of spawned enemies
					spawnedEnemies = append(spawnedEnemies, enemy)
					spawnedPositions = append(spawnedPositions, math.Vector{X: spawnX, Y: spawnY})
				}
			}
		}
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

// growSegment recursively adds connected wall points to a segment
func (s *Spawner) growSegment(wallPoints []worldgen.WallPoint, visited []bool, segment *[]worldgen.WallPoint, currentIdx int) {
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
		// 4.0 is distance squared (2 pixels)
		if distSq < 4.0 && angle < 45.0 {
			*segment = append(*segment, candidate)
			visited[i] = true
			s.growSegment(wallPoints, visited, segment, i)
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

// isTransparent checks if the pixel at the given world coordinates is transparent.
func isTransparent(world *worldgen.GeneratedWorld, x, y float64) bool {
	cell := world.GetCellAt(int(x), int(y))
	if cell == nil || cell.Snippet == nil {
		return true
	}

	localX := x - float64(cell.X*worldgen.CellSize)
	localY := y - float64(cell.Y*worldgen.CellSize)

	if cell.Rotation != 0 {
		angle := -float64(cell.Rotation) * (stdmath.Pi / 180.0)
		cosA := stdmath.Cos(angle)
		sinA := stdmath.Sin(angle)
		cx := float64(worldgen.CellSize) / 2
		cy := float64(worldgen.CellSize) / 2
		relX := localX - cx
		relY := localY - cy
		rotX := relX*cosA - relY*sinA
		rotY := relX*sinA + relY*cosA
		localX = rotX + cx
		localY = rotY + cy
	}

	px := int(localX)
	py := int(localY)

	img := cell.Snippet.Image
	if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
		return true
	}

	_, _, _, a := img.At(px, py).RGBA()
	return a == 0
}

// isSolid checks if the pixel at the given world coordinates is solid (non transparent).
func isSolid(world *worldgen.GeneratedWorld, x, y float64) bool {
	return !isTransparent(world, x, y)
}
