package enemies

// Health mirrors the player Health type and tracks hit points for enemies.
type Health struct {
	Current int
	Max     int
}

// NewHealth returns a Health struct with current equal to max.
func NewHealth(max int) Health {
	return Health{Current: max, Max: max}
}

// Damage subtracts hit points from the enemy.
func (h *Health) Damage(amount int) {
	if amount <= 0 {
		return
	}
	h.Current -= amount
	if h.Current < 0 {
		h.Current = 0
	}
}

// IsDead reports whether the enemy has zero health left.
func (h *Health) IsDead() bool {
	return h.Current <= 0
}
