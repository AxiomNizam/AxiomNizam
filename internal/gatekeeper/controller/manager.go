package controller

import (
	"context"

	"yourapp/internal/gatekeeper/runtime"
)

type Manager struct {
	Runtime *runtime.Engine
}

func NewManager(engine *runtime.Engine) *Manager {
	return &Manager{Runtime: engine}
}

func (m *Manager) Start(ctx context.Context) error {
	return m.Runtime.Run(ctx)
}
