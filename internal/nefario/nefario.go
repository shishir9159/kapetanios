package nefario

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/model"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

type Nefario struct {
	Client *orchestration.Client
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *zap.Logger
	Mu     sync.Mutex
	// cyclic imports
	DBMS      model.DBMS
	Namespace string
}

func NewNefario() *Nefario {

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger := zap.Must(cfg.Build())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	Client, err := orchestration.NewClient()
	if err != nil {
		logger.Error("error creating orchestration client",
			zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Nefario{
		Client:    Client,
		Ctx:       ctx,
		Cancel:    cancel,
		Namespace: os.Getenv("KAPETANIOS_NAMESPACE"),
		Log:       logger,
	}
}
