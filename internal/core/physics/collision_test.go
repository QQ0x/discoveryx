package physics

import (
	"discoveryx/internal/utils/math"
	"testing"
)

// TestCheckCircleCollision tests the circle-circle collision detection function
func TestCheckCircleCollision(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		c1       CircleCollider
		c2       CircleCollider
		expected bool
	}{
		{
			name: "Overlapping circles",
			c1: CircleCollider{
				Position: math.Vector{X: 0, Y: 0},
				Radius:   5,
			},
			c2: CircleCollider{
				Position: math.Vector{X: 3, Y: 4},
				Radius:   5,
			},
			expected: true,
		},
		{
			name: "Touching circles",
			c1: CircleCollider{
				Position: math.Vector{X: 0, Y: 0},
				Radius:   5,
			},
			c2: CircleCollider{
				Position: math.Vector{X: 10, Y: 0},
				Radius:   5,
			},
			expected: true,
		},
		{
			name: "Non-overlapping circles",
			c1: CircleCollider{
				Position: math.Vector{X: 0, Y: 0},
				Radius:   5,
			},
			c2: CircleCollider{
				Position: math.Vector{X: 15, Y: 0},
				Radius:   5,
			},
			expected: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckCircleCollision(tc.c1, tc.c2)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestCheckCircleRectCollision tests the circle-rectangle collision detection function
func TestCheckCircleRectCollision(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		circle        CircleCollider
		rect          RectCollider
		expectedColl  bool
		expectedNormal math.Vector
	}{
		{
			name: "Circle overlapping rectangle",
			circle: CircleCollider{
				Position: math.Vector{X: 0, Y: 0},
				Radius:   5,
			},
			rect: RectCollider{
				Position: math.Vector{X: 3, Y: 0},
				Width:    10,
				Height:   10,
			},
			expectedColl: true,
			expectedNormal: math.Vector{X: -1, Y: 0},
		},
		{
			name: "Circle inside rectangle",
			circle: CircleCollider{
				Position: math.Vector{X: 5, Y: 5},
				Radius:   3,
			},
			rect: RectCollider{
				Position: math.Vector{X: 5, Y: 5},
				Width:    20,
				Height:   20,
			},
			expectedColl: true,
			expectedNormal: math.Vector{X: 0, Y: -1},
		},
		{
			name: "Circle not overlapping rectangle",
			circle: CircleCollider{
				Position: math.Vector{X: 0, Y: 0},
				Radius:   5,
			},
			rect: RectCollider{
				Position: math.Vector{X: 20, Y: 20},
				Width:    10,
				Height:   10,
			},
			expectedColl: false,
			expectedNormal: math.Vector{X: 0, Y: 0},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			collision, normal := CheckCircleRectCollision(tc.circle, tc.rect)
			if collision != tc.expectedColl {
				t.Errorf("Expected collision %v, got %v", tc.expectedColl, collision)
			}

			// Only check normal if collision is expected
			if tc.expectedColl {
				// Check if normal is in the expected direction (don't check exact values)
				if normal.X*tc.expectedNormal.X < 0 || normal.Y*tc.expectedNormal.Y < 0 {
					t.Errorf("Normal vector in wrong direction. Expected direction (%v, %v), got (%v, %v)",
						tc.expectedNormal.X, tc.expectedNormal.Y, normal.X, normal.Y)
				}
			}
		})
	}
}
