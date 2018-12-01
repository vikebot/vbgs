package vbge

import "sync"

// Health is a struct that describes the Health of a player
type Health struct {
	sync.Mutex
	Value int
}

// SafeHealthSynced returns a definite state of player health
func (p *Player) SafeHealthSynced() int {
	p.Health.Lock()
	defer p.Health.Unlock()
	return p.Health.Value
}

// TakeDamage returns the health of the player after taking dmg
func (h *Health) TakeDamage(p *Player) {
	if p.IsDefending {
		h.Value -= defaultDmg / 2
	} else {
		h.Value -= defaultDmg
	}
}

// NewDefaultHealth returns the default health
func NewDefaultHealth() *Health {
	return &Health{
		Value: MaxHealth,
	}
}

// NewHealth accepts an integer value that is returned as
// new health
func NewHealth(health int) *Health {
	return &Health{
		Value: health,
	}
}
