package vbge

import "sync"

// Health is a struct that describes the Health of a player
type Health struct {
	sync.Mutex
	internalValue int
}

// HealthSynced returns a definite state of player health
func (h *Health) HealthSynced() int {
	h.Lock()
	defer h.Unlock()

	return h.internalValue
}

// TakeDamage returns the health of the player after taking dmg
func (h *Health) TakeDamage(p *Player) {
	if p.IsDefending {
		h.internalValue -= defaultDmg / 2
	} else {
		h.internalValue -= defaultDmg
	}
}

// NewDefaultHealth returns the default health
func NewDefaultHealth() *Health {
	return NewHealth(MaxHealth)
}

// NewHealth accepts an integer value that is returned as
// new health
func NewHealth(health int) *Health {
	return &Health{
		internalValue: health,
	}
}
