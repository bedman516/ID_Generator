package main

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"
)

// FactoryConfig - configuration for building the ID facotry
/**
 * NumOfWorkers - how many producers
 * NumOfIDsPerWorker - how much work per producer
 * WorkerID - producer ids, must be unique, will be used as machine id later
 * IDConfig - configuration for ID generator
 *
 *  */
type FactoryConfig struct {
	NumOfWorkers      uint64
	NumOfIDsPerWorker int64
	WorkerIDs         []uint64
	IDConfig          *IDConfig
}

// IDFactory - instance of facotry which produces IDs
/**
 * idQueue - priorityQueue for get ordered IDs
 * mutex - prevent racing between popping and pushing
 * HasWork - check all tasks done or not
 *
 * It's a multiple worker - single consumer factory with limited work.
 * If wanting inf generation, we can add doInfWork function and ignore NumOfIDsPerWorker
 */
type IDFactory struct {
	idQueue       *PriorityQueue
	factoryConfig *FactoryConfig
	mutex         sync.Mutex
	HasWork       sync.WaitGroup
}

// Worker - producer instance
/** generator - ID generator
 * workNum - work number for each worker
 */
type Worker struct {
	generator *Generator
	workNum   int64
}

// NewIDFactory - returns factory instance
func NewIDFactory(w *FactoryConfig) (*IDFactory, error) {
	idQueue := make(PriorityQueue, 0)
	if w.NumOfWorkers != uint64(len(w.WorkerIDs)) && len(w.WorkerIDs) > 0 {
		return nil, errors.New("Please check if workerIDs matches NumOfWorkers")
	}
	heap.Init(&idQueue)
	factory := &IDFactory{
		idQueue:       &idQueue,
		factoryConfig: w,
	}
	factory.HasWork.Add(1)
	return factory, nil
}

// Produce - produces ids
func (m *IDFactory) Produce() {

	workers, err := m.hireWorkers()
	if err != nil {
		fmt.Printf("Failed to produce: %v", err)
		return
	}
	for _, worker := range workers {
		m.HasWork.Add(1)
		go worker.doWork(&m.mutex, m.idQueue, &m.HasWork)
	}
}

// Consume - using PriorityQueue to pop IDs based on timestamp, however, it's roughly sorted
func (m *IDFactory) Consume() {
	defer m.HasWork.Done()
	totalWork := m.factoryConfig.NumOfWorkers * uint64(m.factoryConfig.NumOfIDsPerWorker)
	counter := uint64(0)
	for {
		m.mutex.Lock()
		if len((*m.idQueue)) > 0 {
			id := heap.Pop(m.idQueue)
			fmt.Printf("Timestamp: %d, ID#: %d\n", id.(*ID).timestamp, id.(*ID).IDNum)
			counter++
		}
		m.mutex.Unlock()
		if counter == totalWork {
			fmt.Println("All work done")
			break
		}
	}
}

func (m *IDFactory) hireWorkers() ([]*Worker, error) {
	workers := make([]*Worker, 0)

	for i := 0; uint64(i) < m.factoryConfig.NumOfWorkers; i++ {
		workerID := uint64(i)
		if len(m.factoryConfig.WorkerIDs) > 0 && len(m.factoryConfig.WorkerIDs) > i {
			workerID = m.factoryConfig.WorkerIDs[i]
		}
		config := &IDConfig{
			TimeBit:        m.factoryConfig.IDConfig.TimeBit,
			MachineBit:     m.factoryConfig.IDConfig.MachineBit,
			MachineID:      workerID,
			SequenceBit:    m.factoryConfig.IDConfig.SequenceBit,
			GenerationRate: m.factoryConfig.IDConfig.GenerationRate,
		}
		g, err := NewIDGenerator(config)
		if err != nil {
			return nil, err
		}
		w := &Worker{
			generator: g,
			workNum:   m.factoryConfig.NumOfIDsPerWorker,
		}
		workers = append(workers, w)
	}
	return workers, nil
}

func (w *Worker) doWork(mutex *sync.Mutex, productsQueue *PriorityQueue, HasWork *sync.WaitGroup) {
	defer HasWork.Done()
	for i := 0; int64(i) < w.workNum; i++ {
		id, err := w.generator.GetNextID()
		if err != nil {
			fmt.Printf("Got error when generating id: %v", err)
			break
		}
		mutex.Lock()
		heap.Push(productsQueue, id)
		mutex.Unlock()
	}
}
