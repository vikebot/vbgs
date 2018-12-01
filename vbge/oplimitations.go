package vbge

import "go.uber.org/ratelimit"

// OpLimitations stores ratelimit datastructures for each callable player
// operation
type OpLimitations struct {
	Rotate      ratelimit.Limiter
	Move        ratelimit.Limiter
	Radar       ratelimit.Limiter
	Scout       ratelimit.Limiter
	Environment ratelimit.Limiter
	Watch       ratelimit.Limiter
	Attack      ratelimit.Limiter
	Defend      ratelimit.Limiter
	Health      ratelimit.Limiter
}

// NewOpLimitations returns a new pointer to a new OpLimitation container
// with default time-limitations already set.
func NewOpLimitations() *OpLimitations {
	return &OpLimitations{
		Rotate:      ratelimit.New(rotateThrottle),
		Move:        ratelimit.New(moveThrottle),
		Radar:       ratelimit.New(radarThrottle),
		Scout:       ratelimit.New(scoutThrottle),
		Environment: ratelimit.New(environmentThrottle),
		Watch:       ratelimit.New(watchThrottle),
		Attack:      ratelimit.New(attackThrottle),
		Defend:      ratelimit.New(defendThrottle),
		Health:      ratelimit.New(healthThrottle),
	}
}
