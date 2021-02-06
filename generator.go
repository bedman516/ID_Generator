package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

/**
Default settings
1. Time can be up to 2^38
2. Support machines up to 2^16
3. Rate is 1ms -> 2^10 ids
4. Time unit of genertor: 1ms
**/
const (
	TIMEBIT     = 38
	MACHINEBIT  = 16
	SEQUENCEBIT = 10
	TIMEUNIT    = time.Millisecond
)

// IDConfig - Configuration of ID generator
/**
 * TimeBit - TimeBit defines the maximum value of generator expiration time
 * MachineBit - defines the maximum number of machines
 * SequenceBit - defines the maximum ID that can be generated during the time interval
 * MachineID - Unique ID for machine
 * GenerationRate - defines generation rate,
 * if it's 1s, and seqbit is 16, it means it will generate up to 2^16 IDs in 1s
 */
type IDConfig struct {
	TimeBit        uint16
	MachineBit     uint16
	SequenceBit    uint16
	MachineID      uint64
	GenerationRate time.Duration
}

// Generator - ID generator
/**
 * SeqenceNum - Sequence counter, will roll over when it gets maximum
 * MachineID - Unique ID for machine
 * TimeCounter - Time counter, when it gets 2^timebits, this generator will expire
 * TimeBit - TimeBit defines the maximum value of generator expiration time
 * MachineBit - defines the maximum number of machines
 * SequenceBit - defines the maximum ID that can be generated during the time interval
 * lastTimeStamp - it's used for calculating the time difference
 * lock - used for using same generator to speed up the process
 *
 * */
type Generator struct {
	seqenceNum    uint64
	machineID     uint64
	timeCounter   uint64
	timeBit       uint16
	machineBit    uint16
	sequenceBit   uint16
	rate          int64
	lastTimeStamp int64
	lock          sync.Mutex
}

// ID -
type ID struct {
	IDNum     uint64
	timestamp int64
}

// NewIDGenerator - returns ID generator
func NewIDGenerator(idConfig *IDConfig) (*Generator, error) {
	var err error
	g := &Generator{
		lastTimeStamp: time.Now().UnixNano(),
	}
	// using customized ID config
	ip, err := GetOutboundIP() // use ip address
	if err != nil {
		return nil, err
	}
	defaultConfig := &IDConfig{
		TimeBit:        TIMEBIT,
		MachineBit:     MACHINEBIT,
		SequenceBit:    SEQUENCEBIT,
		GenerationRate: TIMEUNIT,
		MachineID:      uint64(ip),
	}

	if idConfig == nil {
		idConfig = defaultConfig
	}

	if idConfig.TimeBit >= 64 || idConfig.MachineBit >= 64 || idConfig.SequenceBit >= 64 ||
		idConfig.TimeBit+idConfig.MachineBit+idConfig.SequenceBit > 64 {
		return nil, errors.New("Please check the bits you set for config. Single >= 64 or sum > 64 is not allowed")
	}
	// Set machine ID
	if idConfig.MachineBit == 0 {
		g.machineBit = defaultConfig.MachineBit
		g.machineID = idConfig.MachineID
		fmt.Printf("MachineBit can't be 0. It will be set to default machine bit: %d\n", defaultConfig.MachineBit)
	} else if idConfig.MachineBit > 0 && idConfig.MachineID >= uint64(1)<<idConfig.MachineBit {
		g.machineID = defaultConfig.MachineID
		fmt.Printf("MachineID overflows. It will be set to default machine id(ip address): %d\n", defaultConfig.MachineID)
	} else {
		g.machineBit = idConfig.MachineBit
		g.machineID = idConfig.MachineID
	}

	// rate
	if idConfig.GenerationRate <= 0 {
		g.rate = int64(defaultConfig.GenerationRate)
		fmt.Printf("GenerationRate can't be 0. It will be set to default rate: %d\n", defaultConfig.GenerationRate)
	} else {
		g.rate = idConfig.GenerationRate.Nanoseconds()
	}

	//timebits
	if idConfig.TimeBit == 0 {
		g.timeBit = defaultConfig.TimeBit
		fmt.Printf("TimeBit can't be 0. It will be set to default TimeBit: %d\n", defaultConfig.TimeBit)
	} else {
		g.timeBit = idConfig.TimeBit
	}

	//seqbits
	if idConfig.SequenceBit == 0 {
		g.sequenceBit = defaultConfig.SequenceBit
		fmt.Printf("SequenceBit can't be 0. It will be set to default SequenceBit: %d\n", defaultConfig.SequenceBit)
	} else {
		g.sequenceBit = idConfig.SequenceBit
	}

	return g, nil
}

// GetNextID - gets next id value
func (g *Generator) GetNextID() (*ID, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	curTime := time.Now().UnixNano()
	elapsedTime := curTime - g.lastTimeStamp
	// increases time counter after an interval
	if elapsedTime >= g.rate {
		g.timeCounter++
		g.lastTimeStamp = curTime
	}
	// clock check
	if elapsedTime < 0 {
		return nil, errors.New("Please check if the clock is set properly")
	}

	// exceeds time limit
	if g.timeCounter >= (uint64(1) << g.timeBit) {
		return nil, errors.New("Due to time expiration, this generator is no longer valid for generating unique IDs please create a new one to use")
	}

	// seq number needs to roll over if reaching the limit
	if g.seqenceNum >= (uint64(1)<<g.sequenceBit)-1 {
		g.seqenceNum = 0
		if elapsedTime < g.rate {
			// sleep until next round
			time.Sleep(time.Duration(g.lastTimeStamp + g.rate - curTime))
			g.timeCounter++
		}
	} else {
		g.seqenceNum++
	}
	id := &ID{
		IDNum:     g.newID(),
		timestamp: time.Now().UnixNano(),
	}
	return id, nil
}

// newID - generates id using time, machineID and seqNum
func (g *Generator) newID() uint64 {
	timeBits := g.timeCounter << (g.machineBit + g.sequenceBit)
	machineBits := uint64(g.machineID) << g.sequenceBit
	return timeBits | machineBits | g.seqenceNum
}

// GetOutboundIP - Get IP address for default machine ID usage
func GetOutboundIP() (uint16, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return 0, errors.New("Failed to obtain IP address. Please provide your own unique machine ID")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr).IP

	return uint16(localAddr[2]<<8 | localAddr[3]), nil
}
