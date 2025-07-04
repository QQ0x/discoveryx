package physics

import "discoveryx/internal/utils/math"

// Vector is an alias of utils/math.Vector to avoid exposing that package
// directly from the physics layer. It can be extended with additional
// helper methods in the future.
type Vector = math.Vector
