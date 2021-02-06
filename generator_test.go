package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/syncmap"
)

func validateGenerator(g *Generator, config *IDConfig) bool {
	return g.timeBit != config.TimeBit ||
		g.machineBit != config.MachineBit ||
		g.sequenceBit != config.SequenceBit ||
		g.machineID != config.MachineID
}

// TestNewIDGenerator -

func TestNewIDGenerator(t *testing.T) {
	defaultConfig := &IDConfig{
		TimeBit:        38,
		MachineBit:     16,
		MachineID:      1,
		SequenceBit:    10,
		GenerationRate: time.Second,
	}
	config := *defaultConfig
	g, err := NewIDGenerator(&config)

	if err != nil {
		t.Errorf("generator not created: %v", err)
	} else if validateGenerator(g, &config) {
		t.Errorf("config doesn't match, result: %v, expected: %v", g, config)
	}

	// g should be set to default value
	config.SequenceBit = 0
	config.TimeBit = 0
	config.MachineBit = 0
	g, _ = NewIDGenerator(&config)
	if validateGenerator(g, defaultConfig) {
		t.Errorf("config doesn't match, result: %v, expected: %v", g, defaultConfig)
	}

	// sum > 64
	config.SequenceBit = 10
	config.TimeBit = 40
	config.MachineBit = 16
	_, err = NewIDGenerator(&config)
	if err == nil {
		t.Errorf("Generator should not be created due to bits are greater than 64 %v", err)
	}

	// sum < 64
	config.SequenceBit = 8
	config.TimeBit = 15
	config.MachineBit = 5
	_, err = NewIDGenerator(&config)
	if err != nil {
		t.Errorf("Generator should be created successfully, but got %v", err)
	}

}

func TestNewID(t *testing.T) {
	defaultConfig := &IDConfig{
		TimeBit:        38,
		MachineBit:     16,
		MachineID:      1,
		SequenceBit:    10,
		GenerationRate: time.Second,
	}
	gen, err := NewIDGenerator(defaultConfig)
	if err != nil {
		fmt.Printf("generator not created : %+v\n", err)
	}

	id := gen.newID()
	expecedID := (0<<26 | defaultConfig.MachineID<<10 | gen.seqenceNum)
	if id != expecedID {
		t.Errorf("Generated id doesn't match, got: %v , expected: %v", id, expecedID)
	}
	gen.seqenceNum = 1024
	gen.timeCounter = 9

	id = gen.newID()
	expecedID = (9<<26 | 1<<10 | 1024)
	if id != expecedID {
		t.Errorf("Generated id doesn't match, got: %v , expected: %v", id, expecedID)
	}

}

//Test generator for 10s
func TestGetNextID(t *testing.T) {
	var err error
	gen, err := NewIDGenerator(nil)
	if err != nil {
		fmt.Printf("generator not created : %+v\n", err)
	}

	set := make(map[uint64]struct{})

	startTime := time.Now().UnixNano()
	duration := time.Second * 10

	for time.Now().UnixNano()-startTime < int64(duration) {
		id, err := gen.GetNextID()
		if err != nil {
			fmt.Printf("got error when getting next id: %+v\n", err)
		}
		if _, ok := set[id.IDNum]; !ok {
			set[id.IDNum] = struct{}{}
		} else {
			t.Errorf("Got duplicate ID")
		}
	}
}
func generateIDs(g *Generator, wg *sync.WaitGroup, set *syncmap.Map, numIDs int, t *testing.T) {
	defer wg.Done()
	for i := 0; i < numIDs; i++ {
		id, err := g.GetNextID()
		if err != nil {
			fmt.Printf("Got error when generating id: %v", err)
			break
		}
		_, ok := set.Load(id.IDNum)
		if ok {
			t.Errorf("Got duplicated id: %d", id.IDNum)
			break
		}
		set.Store(id.IDNum, struct{}{})
	}
}
func TestDistirbutedGenerators(t *testing.T) {
	set := syncmap.Map{}
	var wg sync.WaitGroup
	numIDs := 1 << 16

	numProcesses := 8
	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		config := &IDConfig{
			TimeBit:        38,
			MachineBit:     6,
			MachineID:      uint64(i),
			SequenceBit:    16,
			GenerationRate: time.Millisecond,
		}
		g, err := NewIDGenerator(config)
		if err != nil {
			fmt.Printf("generator not created : %+v\n", err)
		}
		go generateIDs(g, &wg, &set, numIDs, t)
	}
	wg.Wait()
}

func TestGeneratorInParllel(t *testing.T) {
	set := syncmap.Map{}
	var wg sync.WaitGroup
	numIDs := 1 << 16

	g, err := NewIDGenerator(nil)
	if err != nil {
		fmt.Printf("generator not created : %+v\n", err)
	}
	// having 10 goroutines to use 1 g
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go generateIDs(g, &wg, &set, numIDs, t)
	}
	wg.Wait()
}

// Error cases test
