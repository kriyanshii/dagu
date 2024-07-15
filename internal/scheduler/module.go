package scheduler

import (
	"context"

	"github.com/dagu-dev/dagu/internal/config"
	"github.com/dagu-dev/dagu/internal/engine"
	dagulogger "github.com/dagu-dev/dagu/internal/logger"
	"github.com/dagu-dev/dagu/internal/persistence"
	"github.com/dagu-dev/dagu/internal/persistence/client"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
)

type Params struct {
	fx.In

	Config    *config.Config
	Logger    dagulogger.Logger
	Engine    engine.Engine
	Datastore persistence.DataStores
}

func New(params Params) *Scheduler {
	cfg, _ := config.Load()
	return newScheduler(newSchedulerArgs{
		EntryReader: newEntryReader(newEntryReaderArgs{
			Engine:  params.Engine,
			DagsDir: params.Config.DAGs,
			JobCreator: &jobCreatorImpl{
				WorkDir:    params.Config.WorkDir,
				Engine:     params.Engine,
				Executable: params.Config.Executable,
			},
			Logger: params.Logger,
		}),
		QueueReader: newQueueReader(newQueueReaderArgs{
			QueueDir: params.Config.QueueDir,
			Logger:   params.Logger,
			Datastore: client.NewDataStores(&client.NewDataStoresArgs{
				DAGs:              cfg.DAGs,
				DataDir:           cfg.DataDir,
				SuspendFlagsDir:   cfg.SuspendFlagsDir,
				LatestStatusToday: cfg.LatestStatusToday,
				QueueDir:          cfg.QueueDir,
			}),
		}),
		Logger: params.Logger,
		LogDir: params.Config.LogDir,
	})
}

func LifetimeHooks(lc fx.Lifecycle, a *Scheduler) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) (err error) {
				return a.Start()
			},
			OnStop: func(_ context.Context) error {
				a.Stop()
				return nil
			},
		},
	)
}
