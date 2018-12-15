package ge

import (
	"context"

	"github.com/vikebot/vbgs/pkg/syncc"
)

type CharacterType string

type Character interface {
	CharacterType() CharacterType
	CharacterRotater
	CharacterMover
	CharacterRadarer
	CharacterScouter
	CharacterAttacker
}

type CharacterRotater interface {
	Rotate(ctx context.Context, sr syncc.Request, angle Angle) (ng NotifyGroupLocation)
}

type CharacterMover interface {
	Move(ctx context.Context, sr syncc.Request, dir Direction) (ng NotifyGroupLocation, err error)
}

type CharacterRadarer interface {
	Radar(ctx context.Context, sr syncc.Request) (count int, ng NotifyGroupLocation, err error)
}

type CharacterScouter interface {
	Scout(ctx context.Context, sr syncc.Request, distance int) (count int, ng NotifyGroupLocation, err error)
}

type CharacterAttacker interface {
	DamagePoints() int
	Attack(ctx context.Context, sr syncc.Request, events AttackEvents, changedStats StatsEvent) (enemyHealth int, ng NotifyGroupLocation, err error)
}
