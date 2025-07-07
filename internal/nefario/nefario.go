package nefario

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/model"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	"sync"
)

type Nefario struct {
	client    *orchestration.Client
	ctx       context.Context
	cancel    context.CancelFunc
	log       *zap.Logger
	mu        sync.Mutex
	db        model.DB
	namespace string
}

func (m *Nefario) NewNefario() *Nefario {

	return &Nefario{}
}
