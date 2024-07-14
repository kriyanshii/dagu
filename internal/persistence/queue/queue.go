package queue

import (
	"encoding/json"
	"log"
	"sync"

	// "fmt"
	// "log"
	"os"
	// "strings"
	"errors"
	"path/filepath"

	"github.com/dagu-dev/dagu/internal/dag"
	"github.com/dagu-dev/dagu/internal/persistence/model"
	"github.com/dagu-dev/dagu/internal/util"
)

// type QueueItem struct {
// 	DAGFile string   `json:"dag"`
// 	Params  []string `json:"params"`
// }

type jsonStore struct {
	dir       string
	queueLock sync.Mutex
	Items     []*model.Queue `json:"items"`
}

func New(dirPath string) *jsonStore {
	// dir := filepath.Join(dirPath , "queue.json")
	log.Print("dirPath", dirPath)
	err := os.MkdirAll(dirPath, 0755)
	log.Print("making qeueue", err)
	// _, _ = util.OpenOrCreateFile(dir)
	// log.Printf("queue.json dir:",dirPath)
	return &jsonStore{dir: dirPath}
}

func (store *jsonStore) Create() error {
	// dir := filepath.Join(dirPath , "queue.json")
	queuePath := filepath.Join(store.dir, "queue.json")
	_, err := util.OpenOrCreateFile(queuePath)
	if err != nil {
		return err
	}
	return nil
}

func (store *jsonStore) Save() error {
	queuePath := filepath.Join(store.dir, "queue.json")
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return os.WriteFile(queuePath, data, 0600)
}

// func (store *jsonStore) Open(queue persistence.QueueStore) error{
// 	// is this working?
// 	log.Print("json opened", queue)
// 	return nil
// }

func (store *jsonStore) Load() error {
	queuePath := filepath.Join(store.dir, "queue.json")
	log.Print("reading qeueue")

	data, err := os.ReadFile(queuePath)
	if err != nil {
		if os.IsNotExist(err) {
			store.Items = []*model.Queue{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, store)
}

func (store *jsonStore) Enqueue(d *dag.DAG) error {

	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	log.Print("reading qeueue")

	// log.Print("data:", data)
	store.Items = append(store.Items, &model.Queue{Name: d.Location, Params: d.Params})
	return store.Save()
}

func (store *jsonStore) Dequeue() (*model.Queue, error) {
	store.queueLock.Lock()
	defer store.queueLock.Unlock()
	store.Load()
	if len(store.Items) == 0 {
		return &model.Queue{}, errors.New("queue is empty")
	}
	item := store.Items[0]
	store.Items = store.Items[1:]
	err := store.Save()
	return item, err
}
