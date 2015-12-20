package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

//ChildProcess keeps information about all child processes spawned from the config files
type ChildProcess struct {
	PID     int
	Pname   string
	PPath   string
	IsAlive bool
	Args    string
	//output endpoints for the process are enabled/disable
	EStdErr bool
	EStdOut bool
	EStats  bool
	Proc    *exec.Cmd
	//Responsible for process output redirection
	StdOutWr *io.PipeWriter
	StdOutR  *io.PipeReader
	StdErrWr *io.PipeWriter
	StdErrR  *io.PipeReader
	//Stderr and Stdout storage
	Errors []string
	Output []string
	//time started
	Timestamp time.Time
	//Used to check if process crashed or killed through the api
	KillSwitch bool
	//Number of times restarted
	RestartCount int
	//When process returns, for checking if detachment
	DetachF bool
	lock    sync.RWMutex //for the stdout and stderr storage
}

//Initialize creates the process instance
//redirects stdout and stderr to internal pipes
//starts the process
func (cp *ChildProcess) Initialize(ps process) error {
	wd, bname := getWD(ps.Path)
	if err := os.Chdir(wd); err != nil {
		return err
	}
	if len(ps.Args) < 1 {
		cp.Proc = exec.Command(wd + bname)
	} else {
		log.Println(ps.Args)
		cp.Proc = exec.Command(wd+bname, ps.Args)
	}
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

//getWD  gets the working directory to the process context
func getWD(path string) (string, string) {
	tree := strings.Split(path, "/")
	bname := tree[len(tree)-1]
	tree[len(tree)-1] = ""
	return strings.Join(tree, "/"), bname
}

//Kill stops the process with the KillSwitch flag
func (cp *ChildProcess) Kill() error {
	cp.KillSwitch = true
	return cp.Proc.Process.Kill()
}

//Stats gets the cpu and memory usage of a process
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

//GetErrors gets the current errors from the process stderr
func (cp *ChildProcess) GetErrors() []string {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	return cp.Errors
}

//GetOutput gets the stdout of the process
func (cp *ChildProcess) GetOutput() []string {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	return cp.Output
}

//AppendError stores error info from the stderr of the process
//to the internal slice
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
