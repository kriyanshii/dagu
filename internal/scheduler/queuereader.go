// TO DO - implement a scheduler which
/*
	- checks number of dags running at the same time
	- if numberOfRunningDags < queueLength(from config)
		- it will periodically checks from the queue.json and dequeue first params if
	- this will happen until queue.json is `[]`
*/

package scheduler

import (
	"log"
	"time"

	// "path/filepath"

	"github.com/dagu-dev/dagu/internal/config"
	"github.com/dagu-dev/dagu/internal/engine"
	"github.com/dagu-dev/dagu/internal/persistence/model"

	// "github.com/dagu-dev/dagu/internal/persistence/client"
	"github.com/dagu-dev/dagu/internal/logger"
	"github.com/dagu-dev/dagu/internal/persistence"
)

type newQueueReaderArgs struct {
	QueueDir  string
	Logger    logger.Logger
	Datastore persistence.DataStores
}

type queueReaderImpl struct {
	queueDir string
	// dags          *dag.DAG
	logger logger.Logger
	// engine        engine.Engine
	datastore persistence.DataStores
	// historyStore  persistence.HistoryStore
	queueStore persistence.QueueStore
}

func newQueueReader(args newQueueReaderArgs) *queueReaderImpl {
	// fmt.Print("queue:", queueDir)
	qr := &queueReaderImpl{
		// engine: 	   params.engine,
		queueDir:  args.QueueDir,
		logger:    args.Logger,
		datastore: args.Datastore,
	}

	// log.Print(qr)

	if err := qr.initQueue(); err != nil {
		qr.logger.Error("failed to init queue_reader queue", err)
	}

	return qr
}

func (qr *queueReaderImpl) Start(done chan any) {
	go qr.watchQueue(done)
}

func (qr *queueReaderImpl) watchQueue(done chan any) {
	const checkInterval = 2 * time.Second // Check interval in seconds
	errs := make(chan error)
	ticker := time.NewTicker(checkInterval)
	cfg, _ := config.Load()

	for {
		select {
		case <-ticker.C:
			runFi, err := qr.ReadFileQueue()
			if err != nil {
				errs <- err
				return
			}
			if runFi != nil {
				// fi = newFi
				// for i := 0; i < len(runFi); i++{
				// e := engine.NewFactory(qr.dataStoreFactory, config.Get()).Create()
				// dag := e.GetStatus(runFi.File)
				// log.Printf("%T",runFi[i].Name)
				// }

				e := engine.New(&engine.NewEngineArgs{
					DataStore:  qr.datastore,
					WorkDir:    cfg.WorkDir,
					Executable: cfg.Executable,
				})
				dag, _ := e.GetStatus(runFi.Name)
				e.Start(dag.DAG, "")
				log.Print("nameeee:", runFi.Name)
			}
			if runFi == nil {
				continue
			}
		case <-done:
			return
		}
	}
}

func (qr *queueReaderImpl) ReadFileQueue() (*model.Queue, error) {
	// e := engine.NewFactory(qr.dataStoreFactory, config.Get()).Create()
	// queueLength := cfg.Get().DAGQueueLength
	cfg, _ := config.Load()
	e := engine.New(&engine.NewEngineArgs{
		DataStore:  qr.datastore,
		WorkDir:    cfg.WorkDir,
		Executable: cfg.Executable,
	})
	// log.Print("getrunning status", e.GetrunningStatus)
	// if DAGQueueLength >

	noOfRunningDAGS, _, _ := e.GetrunningStatus()
	log.Print(len(noOfRunningDAGS))
	params, _ := qr.queueStore.Dequeue()
	log.Print("params:", params)
	return params, nil
}

// func execute() {

// }

func (qr *queueReaderImpl) initQueue() error {
	// TODO: do not use the persistence package directly.
	qr.queueStore = qr.datastore.QueueStore()
	err := qr.queueStore.Create()
	if err != nil {
		return err
	}
	return nil
}
