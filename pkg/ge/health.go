package ge

import "github.com/vikebot/vbgs/pkg/syncc"

type Health interface {
	Points() int
	TakeDamage(attacker CharacterAttacker, sr syncc.Request)
}
