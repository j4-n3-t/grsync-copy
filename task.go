package grsync

import (
	"bufio"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
)

// Task is high-level API under rsync
type Task struct {
	rsync *Rsync

	state *State
	log   *Log
}

// State contains information about rsync process
type State struct {
	Remain   int     `json:"remain"`
	Total    int     `json:"total"`
	Speed    string  `json:"speed"`
}

// Log contains raw stderr and stdout outputs
type Log struct {
	Stderr string `json:"stderr"`
}

// State returns inforation about rsync processing task
func (t Task) State() State {
	return *t.state
}

// Log return structure which contains raw stderr and stdout outputs
func (t Task) Log() Log {
	return Log{
		Stderr: t.log.Stderr,
	}
}

// Run starts rsync process with options
func (t *Task) Run() error {
	stderr, err := t.rsync.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	var wg sync.WaitGroup
	go processStderr(&wg, t, stderr)
	// 1?
	wg.Add(2)

	err = t.rsync.Run()
	wg.Wait()

	return err
}

// NewTask returns new rsync task
func NewTask(source, destination string, rsyncOptions RsyncOptions) *Task {
	// Force set required options
	rsyncOptions.HumanReadable = true
	rsyncOptions.Partial = true
	rsyncOptions.Archive = true

	return &Task{
		rsync: NewRsync(source, destination, rsyncOptions),
		state: &State{},
		log:   &Log{},
	}
}

func processStderr(wg *sync.WaitGroup, task *Task, stderr io.Reader) {
	defer wg.Done()

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		task.log.Stderr += scanner.Text() + "\n"
	}
}

func getTaskProgress(remTotalString string) (int, int) {
	const remTotalSeparator = "/"
	const numbersCount = 2
	const (
		indexRem = iota
		indexTotal
	)

	info := strings.Split(remTotalString, remTotalSeparator)
	if len(info) < numbersCount {
		return 0, 0
	}

	remain, _ := strconv.Atoi(info[indexRem])
	total, _ := strconv.Atoi(info[indexTotal])

	return remain, total
}

func getTaskSpeed(data [][]string) string {
	if len(data) < 2 || len(data[1]) < 2 {
		return ""
	}

	return data[1][1]
}
