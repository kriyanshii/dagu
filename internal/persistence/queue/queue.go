package queue

import (
	"encoding/json"
	"sync"

	// "fmt"
	"log"
	"os"

	// "strings"
	// "errors"
	"path/filepath"

	"github.com/dagu-org/dagu/internal/digraph"
	"github.com/dagu-org/dagu/internal/fileutil"
	"github.com/dagu-org/dagu/internal/persistence/model"
)

// type QueueItem struct {
// 	DAGFile string   `json:"dag"`
// 	Params  []string `json:"params"`
// }

type jsonStore struct {
	dir       string
	queueLock sync.Mutex
	Dags      []*model.Queue `json:"dags"`
}

func NewQueueStore(dirPath string) *jsonStore {
	// dir := filepath.Join(dirPath , "queue.json")
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Println("error creating queueDir", err)
	}
	return &jsonStore{dir: dirPath}
}

func (store *jsonStore) Create() error {
	// dir := filepath.Join(dirPath , "queue.json")
	queuePath := filepath.Join(store.dir, "queue.json")
	exists := fileutil.FileExists(queuePath)
	if !exists {
		_, err := fileutil.OpenOrCreateFile(queuePath)
		if err != nil {
			return err
		}
		store.Dags = []*model.Queue{}
		err = store.Save()
		if err != nil {
			log.Print("error saving intial queue: ", err)
		}
	}
	return nil
}

func (store *jsonStore) Save() error {
	queuePath := filepath.Join(store.dir, "queue.json")
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return os.WriteFile(queuePath, data, 0644)
}

func (store *jsonStore) Load() error {
	queuePath := filepath.Join(store.dir, "queue.json")
	data, err := os.ReadFile(queuePath)
	if err != nil {
		if os.IsNotExist(err) {
			store.Dags = []*model.Queue{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, store)
}

func (store *jsonStore) QueueLength() int {
	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	lenQ := len(store.Dags)
	return lenQ
}

func (store *jsonStore) Enqueue(d *digraph.DAG) error {
	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	// log.Print("data:", data)
	store.Dags = append(store.Dags, &model.Queue{Name: d.Location, Params: d.Params})
	return store.Save()
}

func (store *jsonStore) Dequeue() (*model.Queue, error) {
	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	if len(store.Dags) == 0 {
		return nil, nil
	}
	item := store.Dags[0]
	log.Print("dequeue", item)
	store.Dags = store.Dags[1:]
	err := store.Save()
	return item, err
}

func (store *jsonStore) FindJobId(jobid string) (bool, error) {
	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	for i := 0; i < len(store.Dags); i++ {
		if store.Dags[i].Name == jobid {
			store.Dags = append(store.Dags[:i], store.Dags[i+1:]...) // Remove the item
			err := store.Save()
			if err != nil {
				return false, err
			} else {
				return true, nil // Item found and deleted
			}
			// log.Print("jobid", jobid)
		}
	}
	return false, nil
}
