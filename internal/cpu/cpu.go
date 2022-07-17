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
	// reset the File reader to point
	// to the top of the file again
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
		case strings.HasPrefix(line, "intr"):
			_, err := fmt.Sscanf(
				line,
				"intr %d",
				&c.Interrupts,
			)
			if err != nil {
				return err
			}
		case strings.HasPrefix(line, "ctxt"):
			_, err := fmt.Sscanf(
				line,
				"ctxt %d",
				&c.ContextSwitches,
			)
			if err != nil {
				return err
			}
		case strings.HasPrefix(line, "btime"):
			// scan the btime value into int64
			var epoch int64
			_, err := fmt.Sscanf(
				line,
				"btime %d",
				&epoch,
			)
			if err != nil {
				return err
			}

			// since the value represents the boot
			// time expressed in seconds from UNIX epoch,
			// create the time.Time object as such
			c.BootTime = time.Unix(epoch, 0)
		case strings.HasPrefix(line, "processes"):
			_, err := fmt.Sscanf(
				line,
				"processes %d",
				&c.Processes,
			)
			if err != nil {
				return err
			}
		case strings.HasPrefix(line, "procs_running"):
			_, err := fmt.Sscanf(
				line,
				"procs_running %d",
				&c.ProcessesRunning,
			)
			if err != nil {
				return err
			}
		case strings.HasPrefix(line, "procs_blocked"):
			_, err := fmt.Sscanf(
				line,
				"procs_blocked %d",
				&c.ProcessesBlocked,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CPU) readCPULine(s string) error {
	// extract the cpu identifier from the line
	// if a match is found, the first string in slice
	// contains the whole match, while the following
	// strings contain the content of the capture groups
	match := cpuPrefixRegex.FindStringSubmatch(s)

	// return error if there are no matches
	if len(match) == 0 {
		return errNoMatch
	}

	// match[0] is the lines prefix, e.g.
	// "cpu" or "cpu4"
	// cut the prefix from the string, to not
	// deal with the optional number in Load.Scan
	cutLine := strings.ReplaceAll(s, match[0], "")

	// create *Load instance
	load := &Load{}
	// scan the current line
	if err := load.Scan(cutLine); err != nil {
		return err
	}

	// check if there are at least two strings in slice
	// or if match[1] (identifier number) is set
	// otherwise assume that the lines gives the aggregated
	// numbers for the whole cpu package
	if len(match) < 2 || match[1] == "" {
		c.Package = load
		return nil
	}

	// convert the cores id to integer
	id, err := strconv.Atoi(match[1])
	if err != nil {
		return err
	}

	// if the id of the current core
	// is higher than the slices capacity
	// create enough slice with a high
	// enough capacity, copy all elements
	// from the current c.Cores slice
	// and overwrite the current c.Cores slice
	if cap(c.Cores) < id+1 {
		tmp := make([]*Load, id+1)
		copy(tmp, c.Cores)
		c.Cores = tmp
	}

	c.Cores[id] = load

	return nil
}

// String (CPU) implements Stringer interface
// and returns the CPU instances string representation
func (c CPU) String() string {
	return fmt.Sprintf(
		"package: %v, cores: %v, interrupts: %d, context switches: %d, boot time: %s, processes: %d, processes running: %d, processes blocked: %d",
		c.Package, c.Cores, c.Interrupts,
		c.ContextSwitches, c.BootTime, c.Processes,
		c.ProcessesRunning, c.ProcessesBlocked,
	)
}
