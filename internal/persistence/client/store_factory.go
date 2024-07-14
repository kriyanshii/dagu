package client

import (
	"os"

	"github.com/dagu-dev/dagu/internal/persistence"
	"github.com/dagu-dev/dagu/internal/persistence/jsondb"
	"github.com/dagu-dev/dagu/internal/persistence/local"
	"github.com/dagu-dev/dagu/internal/persistence/local/storage"
	"github.com/dagu-dev/dagu/internal/persistence/queue"
)

var _ persistence.DataStores = (*dataStores)(nil)

type dataStores struct {
	historyStore persistence.HistoryStore
	dagStore     persistence.DAGStore
	queueStore   persistence.QueueStore

	dags              string
	dataDir           string
	queueDir          string
	suspendFlagsDir   string
	latestStatusToday bool
}

type NewDataStoresArgs struct {
	DAGs              string
	DataDir           string
	SuspendFlagsDir   string
	LatestStatusToday bool
	QueueDir          string
}

func NewDataStores(args *NewDataStoresArgs) persistence.DataStores {
	dataStoreImpl := &dataStores{
		dags:              args.DAGs,
		dataDir:           args.DataDir,
		suspendFlagsDir:   args.SuspendFlagsDir,
		latestStatusToday: args.LatestStatusToday,
		queueDir:          args.QueueDir,
	}
	_ = dataStoreImpl.InitDagDir()
	return dataStoreImpl
}

func (f *dataStores) InitDagDir() error {
	_, err := os.Stat(f.dags)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(f.dags, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (f *dataStores) HistoryStore() persistence.HistoryStore {
	// TODO: Add support for other data stores (e.g. sqlite, postgres, etc.)
	if f.historyStore == nil {
		f.historyStore = jsondb.New(
			f.dataDir, f.dags, f.latestStatusToday)
	}
	return f.historyStore
}

func (f *dataStores) DAGStore() persistence.DAGStore {
	if f.dagStore == nil {
		f.dagStore = local.NewDAGStore(&local.NewDAGStoreArgs{Dir: f.dags})
	}
	return f.dagStore
}

func (f *dataStores) FlagStore() persistence.FlagStore {
	return local.NewFlagStore(storage.NewStorage(f.suspendFlagsDir))
}

func (f dataStores) QueueStore() persistence.QueueStore {
	if f.queueStore == nil {
		f.queueStore = queue.New(f.queueDir)
	}
	return f.queueStore
}
