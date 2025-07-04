package player

// Health represents hit points for an entity. It tracks the current and
// maximum health values and provides helpers for damage application.
type Health struct {
	Current int // Current hit points
	Max     int // Maximum hit points
}

// NewHealth creates a new Health struct with the given maximum value.
// The current value is initialised to the maximum.
func NewHealth(max int) Health {
	return Health{Current: max, Max: max}
}

// Damage reduces the current health by the specified amount. Health will not
// drop below zero.
func (h *Health) Damage(amount int) {
	if amount <= 0 {
		return
	}
	h.Current -= amount
	if h.Current < 0 {
		h.Current = 0
	}
}

// IsDead returns true if the entity has no remaining health.
func (h *Health) IsDead() bool {
	return h.Current <= 0
}
