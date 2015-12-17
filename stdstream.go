package main

import (
	"bufio"
	"time"
)

func LogStdOut(cp *ChildProcess) {
	stdOutScanner := bufio.NewScanner(cp.StdOutR)
	for stdOutScanner.Scan() {
		cp.AppendOutput(stdOutScanner.Text())
	}
}

func LogStdErr(cp *ChildProcess) {
	stdErrScanner := bufio.NewScanner(cp.StdErrR)
	for stdErrScanner.Scan() {
		cp.AppendError(stdErrScanner.Text())
	}
}

func SendPLogs(pl plog, cp *ChildProcess) {
	ticker := time.NewTicker(time.Duration(pl.Interval) * time.Minute)
	stop := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			if pl.Mail == "on" {
				//send mail
			}
			if pl.Twitter == "on" {
				//post to twitter
			}
		case <-stop:
			ticker.Stop()
		}
	}
}
func CrashReport(ps process) {
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
func StreamOutToTwitter(cp *ChildProcess) {}
func AirDirtyLaundry(cp *ChildProcess)    {}
