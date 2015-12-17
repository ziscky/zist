package main

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

//ChildProcess keeps information about all child processes spawned from the config files
type ChildProcess struct {
	PID          int
	Pname        string
	PPath        string
	IsAlive      bool
	EStdErr      bool
	EStdOut      bool
	EStats       bool
	Proc         *exec.Cmd
	StdOutWr     *io.PipeWriter
	StdOutR      *io.PipeReader
	StdErrWr     *io.PipeWriter
	StdErrR      *io.PipeReader
	Errors       []string
	Output       []string
	Timestamp    time.Time
	KillSwitch   bool
	RestartCount int
	DetachF      bool
	lock         sync.RWMutex
}

func (cp *ChildProcess) Initialize(target string) error {
	cp.Proc = exec.Command(target)
	cp.StdOutR, cp.StdOutWr = io.Pipe()
	cp.StdErrR, cp.StdErrWr = io.Pipe()
	cp.Proc.Stdout = cp.StdOutWr
	cp.Proc.Stderr = cp.StdErrWr
	if err := cp.Proc.Start(); err != nil {
		return err
	}
	cp.PID = cp.Proc.Process.Pid
	cp.Timestamp = time.Now()
	return nil
}

func (cp *ChildProcess) Kill() error {
	cp.KillSwitch = true
	return cp.Proc.Process.Kill()
}

func (cp *ChildProcess) Stats() (map[string]string, error) {
	stats, err := exec.Command("ps", "-p", strconv.Itoa(cp.PID), "-o", "%cpu,%mem").Output()
	if err != nil {
		return map[string]string{}, err
	}
	statsStr := strings.Trim(strings.SplitAfter(string(stats), "\n")[1], " ")
	l := strings.SplitAfter(statsStr, " ")

	return map[string]string{
		"cpu": l[0],
		"mem": strings.Replace(l[2], "\n", "", -1),
	}, nil
}

func (cp *ChildProcess) GetErrors() []string {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	return cp.Errors
}
func (cp *ChildProcess) GetOutput() []string {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	return cp.Output
}
func (cp *ChildProcess) AppendError(errorStr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	cp.Errors = append(cp.Errors, errorStr)
}
func (cp *ChildProcess) AppendOutput(outputStr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	cp.Output = append(cp.Output, outputStr)
}
func (cp *ChildProcess) Detach() error {
	cp.DetachF = true
	cp.Kill()
	RemoveProcess(cp)
	cmd := exec.Command(cp.PPath, "&")
	return cmd.Start()
}
