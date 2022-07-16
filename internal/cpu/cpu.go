package cpu

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errNoMatch = fmt.Errorf("no match found in string")

	procFile       = path.Join("/", "proc", "stat")
	cpuPrefixRegex = regexp.MustCompile("cpu([0-9]*)")
)

// Load struct represents CPU or Core load
type Load struct {
	User           int64
	Nice           int64
	System         int64
	Idle           int64
	IOWait         int64
	Interrupts     int64
	SoftInterrupts int64
}

// Scan (Load) reads a the load status from string
func (l *Load) Scan(s string) error {
	_, err := fmt.Sscanf(
		s, "%d %d %d %d %d %d %d",
		&l.User, &l.Nice, &l.System, &l.Idle,
		&l.IOWait, &l.Interrupts, &l.SoftInterrupts,
	)
	if err != nil {
		return err
	}

	return nil
}

// CPU struct represents contents of the /proc/stats file
type CPU struct {
	Package          *Load
	Cores            []*Load
	Interrupts       int64
	ContextSwitches  int64
	BootTime         time.Time
	Processes        int64
	ProcessesRunning int64
	ProcessesBlocked int64

	fp *os.File
}

// New (CPU) creates a new CPU instance
func New() (*CPU, error) {
	fp, err := os.Open(procFile)
	if err != nil {
		return nil, err
	}

	return &CPU{
		fp: fp,

		Cores: make([]*Load, 0),
	}, nil
}

// Close (CPU) closes the CPU instance
func (c *CPU) Close() error {
	return c.fp.Close()
}

// Update (CPU) updates the values
// in the current CPU instance by reading
// them from file
func (c *CPU) Update() error {
	_, err := c.fp.Seek(0, 0)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(c.fp)
	sc.Split(bufio.ScanLines)

	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "cpu"):
			err := c.readCPULine(line)
			if err != nil {
				if errors.Is(err, errNoMatch) {
					continue
				}

				return err
			}
		}
	}

	return nil
}

func (c *CPU) readCPULine(s string) error {
	match := cpuPrefixRegex.FindStringSubmatch(s)

	if len(match) == 0 {
		return errNoMatch
	}

	cpuString := match[0]
	cutLine := strings.ReplaceAll(s, cpuString, "")

	load := &Load{}
	if err := load.Scan(cutLine); err != nil {
		return err
	}

	if len(match) > 2 {
		idString := match[1]
		id, err := strconv.Atoi(idString)
		if err != nil {
			return err
		}

		if cap(c.Cores) < id+1 {
			c.Cores = make([]*Load, id+1)
		}

		c.Cores[id] = load
	} else {
		c.Package = load
	}

	return nil
}
