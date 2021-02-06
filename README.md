ID_Generator
=========

[![Build Status](https://travis-ci.org/sony/sonyflake.svg?branch=master)](https://travis-ci.org/sony/sonyflake)
[![Coverage Status](https://coveralls.io/repos/sony/sonyflake/badge.svg?branch=master&service=github)](https://coveralls.io/github/sony/sonyflake?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/sony/sonyflake)](https://goreportcard.com/report/github.com/sony/sonyflake)

ID_Generator is a distributed unique ID generator inspired by [Snowflake](https://blog.twitter.com/2010/announcing-snowflake) and [Sonyflake](https://github.com/sony/sonyflake). 

This generator can support customized ID settings for different requirements. The default settings contains:

    - 40 bits for time, which gives approximately 348 years lifetime
    - 16 bits for machine ID, which supports up to 2^16 machines
    - 10 bits for sequence number
    - 10 ms for generation speed



Usage
-----

The function NewIDGenerator creates a new Generator instance.

```go
func NewIDGenerator(idConfig *IDConfig) (*Generator, error) 
```

You can configure Sonyflake by the struct Settings:

```go
type IDConfig struct {
	TimeBit        uint16
	MachineBit     uint16
	SequenceBit    uint16
	MachineID      uint64
	GenerationRate time.Duration
}
```


 * TimeBit - TimeBit defines the maximum value of generator expiration time.
 * MachineBit - defines the maximum number of machines.
 * SequenceBit - defines the maximum ID that can be generated during the time interval.
 * MachineID - Unique ID for machine.
 * GenerationRate - defines generation rate. If it's 1s, and SequenceBit is 16, which means it will generate up to 2^16 IDs in 1s


<br>
To get a new unique ID, just call GetNextID.

```go
func (g *Generator) GetNextID() (*ID, error)
```

Here is an example to generate roughly time-ordered IDs:

You can configure the factory example by the FactoryConfig and you will also need to configure the IDConfig as mentioned above. <br>
Please be aware of that the workerID must be unique because it will be used as machineID in generator.

```go
type FactoryConfig struct {
	NumOfWorkers      uint64
	NumOfIDsPerWorker int64
	WorkerIDs         []uint64
	IDConfig          *IDConfig
}
```

To get a factory instance, just call  NewIDFactory
```go
func NewIDFactory(w *FactoryConfig) (*IDFactory, error) 
```
After you call NewIDFactory, it will increase waitGroup counter of factory instance by 1.

To print IDs,  you will need to run Produce and Consume methods. IDs will be printed in Consume method, and feel free to modify it for your own need.
```go
go factory.Produce()
go factory.Consume()
factory.HasWork.Wait()
```
License
-------

The MIT License (MIT)

See [LICENSE](https://github.com/sony/sonyflake/blob/master/LICENSE) for details.