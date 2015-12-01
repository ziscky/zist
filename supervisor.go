package main

import (
	"bufio"
	"flag"
	"fmt"
	"sync"
	"time"
)

var (
	activeProcesses map[int]ChildProcess
	procLock        sync.RWMutex
)

func AddProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	activeProcesses[cp.PID] = *cp
}

func RemoveProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	delete(activeProcesses, cp.PID)
}

func RegisterProcess(ps process) error {
	cp := new(ChildProcess)
	if err := cp.Initialize(ps.Path); err != nil {
		return err
	}
	cp.Pname = ps.Pname
	//Start logging the process stdout
	go func() {
		stdOutScanner := bufio.NewScanner(cp.StdOutR)
		for stdOutScanner.Scan() {
			cp.AppendOutput(stdOutScanner.Text())
		}
	}()
	//Start logging the process stderr
	go func() {
		stdErrScanner := bufio.NewScanner(cp.StdErrR)
		for stdErrScanner.Scan() {
			cp.AppendError(stdErrScanner.Text())
		}
	}()
	if len(ps.Plogs.Plogs) > 0 {
		for _, plog := range ps.Plogs.Plogs {
			go func() {
				ticker := time.NewTicker(time.Duration(plog.Interval) * time.Minute)
				stop := make(chan struct{})
				for {
					select {
					case <-ticker.C:
						if plog.Mail == "on" {
							//send mail
						}
						if plog.Twitter == "on" {
							//post to twitter
						}
					case <-stop:
						ticker.Stop()
					}
				}
			}()
		}
	}
	cp.EStats = ps.Outputs.Stats.Web == "on"
	cp.EStdErr = ps.Outputs.StdErr.Web == "on"
	cp.EStdOut = ps.Outputs.StdOut.Web == "on"

	if ps.Outputs.StdOut.Twitter == "on" {
		go func() {
			//post to twitter after interval
		}()
	}
	if ps.Outputs.StdErr.Twitter == "on" {
		go func() {
			//post to twitter after interval
		}()
	}

	AddProcess(cp)
	if err := cp.Proc.Wait(); err != nil {
		return err
	}
	if !cp.KillSwitch && ps.Restart == "on" { //restart process
		if ps.CrashReport.Enable {
			if len(ps.CrashReport.Mail.Rcps) > 0 {
				if ps.CrashReport.Mail.Body == "log" {

				}
				if ps.CrashReport.Mail.Body == "stderr" {
				}
				if ps.CrashReport.Mail.Body == "stdout" {
				}
				if ps.CrashReport.Mail.Body == "combined" {
				}
				//send body
			}
			if ps.CrashReport.Twitter.Message != "" {
				//tweet message and hashtag
			}
		}
		RemoveProcess(cp)
		return RegisterProcess(ps)
	}
	return nil
}

var (
	appConf *App
	err     error
)

func main() {
	confPath := flag.String("conf", "", "-conf=/home/conf.xml")

	appConf, err = ParseXMLDirectives(*confPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, ps := range appConf.Ps.Pss {
		go RegisterProcess(ps)
	}
	// http.HandleFunc("/{token}", handler)
	// http.HandleFunc("/{token}/{pid}/stats", handler)
	// http.HandleFunc("/{token}/{pid}/stdout", handler)

	// http.HandleFunc("/{token}/{pid}/stderr", handler)
	// http.HandleFunc("/{token}/{pid}/kill", handler)
	// http.HandleFunc("/{token}/{pid}/start", handler)
}
