package main

import (
	"fmt"
	"time"
)

func main() {
	idConfig := &IDConfig{
		TimeBit:        38,
		MachineBit:     6,
		SequenceBit:    16,
		GenerationRate: time.Millisecond,
	}
	ftConfig := &FactoryConfig{
		NumOfWorkers:      5,
		NumOfIDsPerWorker: 1 << 15,
		IDConfig:          idConfig,
	}
	factory, err := NewIDFactory(ftConfig)
	if err != nil {
		fmt.Printf("Factory failed: %v\n", err)
	}
	go factory.Produce()
	go factory.Consume()
	factory.HasWork.Wait()
}
