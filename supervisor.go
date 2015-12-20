package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
)

var (
	activeProcesses map[int]*ChildProcess
	procLock        sync.RWMutex
)

//TODO: flag for process threshold-> min start time to avoid restarting failing procs
//TODO: add support to check if process is using any ports for listening/connecting

func AddProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	activeProcesses[cp.PID] = cp
}

func RemoveProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	delete(activeProcesses, cp.PID)
}

func RegisterProcess(ps process, rcount int) error {
	cp := new(ChildProcess)
	//log.Println(ps.Args)
	if _, err := os.Stat(ps.Path); os.IsNotExist(err) {
		log.Println(ps.Path, " does not exist")
		return err
	}
	if err := cp.Initialize(ps); err != nil {
		log.Println("initialize", err)
		return err
	}
	cp.Pname = ps.Pname
	cp.PPath = ps.Path
	cp.Args = ps.Args
	cp.IsAlive = true
	cp.RestartCount = rcount

	//web statistics settings
	cp.EStats = ps.Outputs.Stats.Web == "on"
	cp.EStdErr = ps.Outputs.StdErr.Web == "on"
	cp.EStdOut = ps.Outputs.StdOut.Web == "on"

	//Start logging the process stdout
	go LogStdOut(cp)
	//Start logging the process stderr
	go LogStdErr(cp)

	if len(ps.Plogs.Plogs) > 0 {
		for _, plog := range ps.Plogs.Plogs {
			go SendPLogs(plog, cp)
		}
	}

	if ps.Outputs.StdOut.Twitter == "on" {
		go StreamOutToTwitter(cp) //still in dev
	}

	if ps.Outputs.StdErr.Twitter == "on" {
		go AirDirtyLaundry(cp) //still in dev
	}

	AddProcess(cp)
	fmt.Println("[*]", cp.Pname, "started successfully.")
	if err := cp.Proc.Wait(); err != nil {
		fmt.Println("[*]", cp.Pname, "Non zero exit: ", err)
	}

	if !cp.KillSwitch {
		if ps.Restart == "on" { //restart process
			if ps.CrashReport.Enable {
				CrashReport(ps)
			}
			RemoveProcess(cp)
			return RegisterProcess(ps, cp.RestartCount+1)
		}
		RemoveProcess(cp)
	}
	if !cp.DetachF {
		if _, ok := activeProcesses[cp.PID]; ok {
			activeProcesses[cp.PID].IsAlive = false
		}
	}
	return nil
}

//AttachProcess invokes ps -ef and greps
//for  the path then launching it the normal process invocation structure
func AttachProcess(rw http.ResponseWriter, r *http.Request) {
	//	pid := r.FormValue("pid")
}

var (
	appConf *App
	err     error
)

func main() {
	confPath := flag.String("conf", "", "-conf=/path/to/conffile")
	keyfile := flag.String("keyfile", "", "-keyfile=path/to/keyfile")
	certfile := flag.String("certfile", "", "-certfile=path/to/certfile")
	flag.Parse()

	appConf, err = ParseXMLDirectives(*confPath)
	if err != nil {
		log.Println(err)
		return
	}
	activeProcesses = make(map[int]*ChildProcess)

	for _, ps := range appConf.Ps.Pss {
		go RegisterProcess(ps, 0)
		//log.Println(ps.Pname)
	}

	router := mux.NewRouter()
	router.HandleFunc("/{token}", CheckToken(Default))
	router.HandleFunc("/{token}/{pid}/stats", CheckToken(WithProcess(Stats)))
	router.HandleFunc("/{token}/{pid}/kill", CheckToken(WithProcess(Kill)))
	router.HandleFunc("/{token}/{pid}/start", CheckToken(WithProcess(Start)))
	router.HandleFunc("/{token}/{pid}/restart", CheckToken(WithProcess(Restart)))
	router.HandleFunc("/{token}/{pid}/stdout", CheckToken(WithProcess(StdOut)))
	router.HandleFunc("/{token}/{pid}/stderr", CheckToken(WithProcess(StdErr)))
	router.HandleFunc("/{token}/{pid}/detach", CheckToken(WithProcess(Detach)))
	//ability to detach process
	switch appConf.Ep.Protocol {
	case "http":
		http.ListenAndServe(":"+appConf.Ep.Port, router)
	case "https": //key file and certificate file needed
		http.ListenAndServeTLS(":"+appConf.Ep.Port, *certfile, *keyfile, router)
	default:
		log.Println("Invalid protocol:", appConf.Ep.Protocol)
	}
}
