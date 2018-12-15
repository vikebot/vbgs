package ge

import (
	"context"
	"github.com/vikebot/vbgs/pkg/syncc"
	"go.uber.org/zap"
)

type CharacterBaseExtender interface {
	Init(cb *CharacterBase, log *zap.Logger)
	CharacterType() CharacterType
}

type Archer struct {
	cb  *CharacterBase
	log *zap.Logger
}

func (Archer) Rotate(ctx context.Context, sr syncc.Request, angle Angle) (ng NotifyGroupLocation) {
	panic("implement me")
}

func (a *Archer) Init(cb *CharacterBase, log *zap.Logger) {
	a.cb = cb
	a.log = log
}

func (Archer) CharacterType() CharacterType {
	return "archer"
}

type CharacterBase struct {
	ec  *EntityCore
	ext CharacterBaseExtender
}

func (cb *CharacterBase) CharacterType() CharacterType {
	if cb.ext != nil {
		return cb.ext.CharacterType()
	}

	return "base"
}

func (cb *CharacterBase) Rotate(ctx context.Context, sr syncc.Request, angle Angle) (ng NotifyGroupLocation) {
	// Call the characters Rotate override
	if r, ok := cb.ext.(CharacterRotater); ok {
		return r.Rotate(ctx, sr, angle)
	}

	// Perform basic rotate
	panic("implement me")
}

func (CharacterBase) Move(ctx context.Context, sr syncc.Request, dir Direction) (ng NotifyGroupLocation, err error) {
	panic("implement me")
}

func (CharacterBase) Radar(ctx context.Context, sr syncc.Request) (count int, ng NotifyGroupLocation, err error) {
	panic("implement me")
}

func (CharacterBase) Scout(ctx context.Context, sr syncc.Request, distance int) (count int, ng NotifyGroupLocation, err error) {
	panic("implement me")
}

func (CharacterBase) DamagePoints() int {
	panic("implement me")
}

func (CharacterBase) Attack(ctx context.Context, sr syncc.Request, events AttackEvents, changedStats StatsEvent) (enemyHealth int, ng NotifyGroupLocation, err error) {
	panic("implement me")
}
