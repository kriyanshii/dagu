// TO DO - implement a scheduler which
/*
- checks number of dags running at the same time
- if numberOfRunningDags < queueLength(from config)
- it will periodically checks from the queue.json and dequeue first params if
- this will happen until queue.json is `[]`
*/

package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"errors"

	"github.com/dagu-org/dagu/internal/client"
	"github.com/dagu-org/dagu/internal/config"
	"github.com/dagu-org/dagu/internal/persistence/model"

	"github.com/dagu-org/dagu/internal/logger"
	"github.com/dagu-org/dagu/internal/persistence"
	"github.com/dagu-org/dagu/internal/persistence/queue"
	"github.com/dagu-org/dagu/internal/persistence/stats"
)

// var _ queueReader = (*queueReaderImpl)(nil)
var (
	errQueueEmpty = errors.New("queue empty")
)

type newQueueReaderArgs struct {
	QueueDir string
	Client   client.Client
}

type queueReaderImpl struct {
	queueDir   string
	logger     logger.Logger
	queueStore persistence.QueueStore
	client     client.Client
}

type StartOptions struct {
	Params           string
	Quiet            bool
	FromWaitingQueue bool
}

type QueueReader interface {
	Start(ctx context.Context, done chan any) error
	ReadFileQueue(ctx context.Context) ([]*model.Queue, error)
}

func NewQueueReader(queueDir string, client client.Client) *queueReaderImpl {
	// fmt.Print("queue:", queueDir)
	qr := &queueReaderImpl{
		queueDir: queueDir,
		client:   client,
	}

	return qr
}

func (qr *queueReaderImpl) Start(ctx context.Context, done chan any) error {
	qr.queueStore = queue.NewQueueStore(qr.queueDir)
	if err := qr.initQueue(); err != nil {
		return fmt.Errorf("failed to init queue", err)
	}
	go qr.watchQueue(ctx, done)
	return nil
}

func (qr *queueReaderImpl) watchQueue(ctx context.Context, done chan any) {
	log.Print("queue being watched")
	const checkInterval = 2 * time.Second // Check interval in seconds
	errs := make(chan error)
	ticker := time.NewTicker(checkInterval)
	// cfg, _ := config.Load()

	for {
		select {
		case <-ticker.C:
			runFi, err := qr.ReadFileQueue(ctx)
			if err != nil {
				errs <- err
				return
			}
			if len(runFi) != 0 {
				for i := 0; i < len(runFi); i++ {
					log.Print("dags readFileQueue:", runFi[i])
					go qr.execute(ctx, runFi[i])
				}
			}
			if runFi == nil {
				continue
			}
		case <-done:
			return
		}
	}
}

// edge case - where noOfDags in queue is less then queueLength
func (qr *queueReaderImpl) ReadFileQueue(ctx context.Context) ([]*model.Queue, error) {
	var params []*model.Queue
	cfg, _ := config.Load()
	stats := stats.NewStatsStore(cfg.Paths.StatsDir)

	queueLength := cfg.DAGQueueLength
	// log.Println("queueLength: ", queueLength)
	noOfRunningDAGS, _ := stats.GetRunningDags()
	// log.Println("noOfRunningDAGS: ", noOfRunningDAGS)

	if queueLength > noOfRunningDAGS {
		for i := 0; i < queueLength-noOfRunningDAGS; i++ {
			DAGparam, err := qr.queueStore.Dequeue()
			// if the queue is empty
			if DAGparam == nil {
				return params, nil
			}
			// if there is any error reading queue.json
			if err != nil {
				logger.Error(ctx, "error reading queue", "error", err)
				return nil, nil
			}
			log.Print("para", DAGparam)
			params = append(params, DAGparam)
		}
		return params, nil
	} else {
		return nil, nil
	}
}

func (qr *queueReaderImpl) execute(ctx context.Context, dagFile *model.Queue) {

	dag, _ := qr.client.GetDAG(ctx, dagFile.Name)
	if err := qr.client.Start(ctx, dag, client.StartOptions{
		Quiet:            false,
		FromWaitingQueue: true,
	}); err != nil {
		logger.Error(ctx, "error starting the dag from queue:", "error", err)
	}
	defer log.Print("executing from queue", dagFile.Name)

}

func (qr *queueReaderImpl) initQueue() error {
	// TODO: do not use the persistence package directly.
	err := qr.queueStore.Create()
	if err != nil {
		return err
	}
	return nil
}
