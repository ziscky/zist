/*
Copyright (C) 2016  Eric Ziscky

    This program is free software; you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation; either version 2 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License along
    with this program; if not, write to the Free Software Foundation, Inc.,
    51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"encoding/json"
	"fmt"
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
func (cp *ChildProcess) Initialize(ps Job,numrestarts int) error {
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
    cp.RestartCount += numrestarts
    if ps.Workingdir != ""{
        cp.Proc.Dir = ps.Workingdir
    }else{
        cp.Proc.Dir = wd
    }
	if err := cp.Proc.Start(); err != nil {
		return err
	}
	cp.PID = cp.Proc.Process.Pid
	cp.Timestamp = time.Now()
    cp.IsAlive = true
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

//AppendOutput stores stdout info from the stdout of the process
//to the internal slice
func (cp *ChildProcess) AppendOutput(outputStr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	cp.Output = append(cp.Output, outputStr)
}

//ClearErrorBuff clears the process error buffer with an option to store it to a file
func (cp *ChildProcess) ClearErrorBuff(store int) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if store > 0 {
		file, err := os.OpenFile("stderr", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("[*] Error writing error buffer to file. Ensure correct permissions")
		}
		errbytes, _ := json.Marshal(cp.Errors)
		if _, err := io.WriteString(file, string(errbytes)); err != nil {
			fmt.Println("[*] Error writing error buffer to file. Ensure correct permissions")
		}
	}
	cp.Errors = cp.Errors[:0]
}

//ClearStdoutBuff clears the process stdout buffer with an option to store it to a file
func (cp *ChildProcess) ClearStdoutBuff(store int) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if store > 0 {
		file, err := os.Open("stdout")
		if err != nil {
			fmt.Println("[*] Error writing stdout buffer to file. Ensure correct permissions")
		}
		errbytes, _ := json.Marshal(cp.Errors)
		if _, err := io.WriteString(file, string(errbytes)); err != nil {
			fmt.Println("[*] Error writing stdout buffer to file. Ensure correct permissions")
		}
	}
	cp.Output = cp.Output[:0]
}

//Detach disowns the child process
func (cp *ChildProcess) Detach() error {
	cp.DetachF = true
	cp.Kill()
	RemoveProcess(cp)
	cmd := exec.Command(cp.PPath, "&")
	return cmd.Start()
}
