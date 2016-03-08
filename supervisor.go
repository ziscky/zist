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
	"flag"
	"fmt"
	"log"
    "net"
	"net/http"
    "net/rpc"
    "encoding/base64"
    "crypto/rand"
	"os"
	"sync"
    "time"
	"github.com/gorilla/mux"
	"strconv"
)

var (
	activeProcesses map[int]*ChildProcess
	procLock        sync.RWMutex
)

//TODO: render last zistd output to zistcl via rpc for easy error checking

//AddProcess adds a process to the process map
func AddProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	activeProcesses[cp.PID] = cp
}

//RemoveProcess removes a process from the process map
func RemoveProcess(cp *ChildProcess) {
	procLock.Lock()
	defer procLock.Unlock()
	delete(activeProcesses, cp.PID)
}

//RegisterProcess initializes a process and adds it to the proccess map
func RegisterProcess(ps Job, rcount int) error {
	cp := new(ChildProcess)
	log.Println(ps.Name)
	if _, err := os.Stat(ps.Path); os.IsNotExist(err) {
		log.Println(ps.Path, " does not exist")
		return err
	}
    
	if err := cp.Initialize(ps,rcount); err != nil {
		log.Println("initialize", err)
		return err
	}
	cp.Pname = ps.Name
	cp.PPath = ps.Path
	cp.Args = ps.Args
	cp.IsAlive = true
	//web statistics settings
    if appConf.Web{
        cp.EStats = ps.Web
        cp.EStdErr = ps.Web
        cp.EStdOut = ps.Web
    }
	//Start logging the process stdout
	go LogStdOut(cp)
	//Start logging the process stderr
	go LogStdErr(cp)


	AddProcess(cp)
	fmt.Println("[*]", cp.Pname, "started successfully.")
	if err := cp.Proc.Wait(); err != nil {
		fmt.Println("[*]", cp.Pname, "Non zero exit: ", err)
	}

	if !cp.KillSwitch {
		if ps.Restart { //restart process
			//crash report
            if time.Since(cp.Timestamp).Seconds() > 10{
                RemoveProcess(cp)
                return RegisterProcess(ps, cp.RestartCount+1)
            }else{
                log.Println(cp.Pname,"is exiting too quick. Backing off. Start Explicitly")
                if _, ok := activeProcesses[cp.PID]; ok {
			         activeProcesses[cp.PID].IsAlive = false
		        }
                return nil
            }
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
	err     error
)


func parseCommand() bool{
    if len(os.Args) > 1{
        if os.Args[1] == "install"{
            if err := install(); err != nil{
                os.Exit(-1)
            }
            fmt.Println("[*] Zist has been succesfully installed")
            return true
        }
        if os.Args[1] == "generate"{
            fmt.Println("[*] Safe Token: ",generateToken())
            return true
        }
    }
    return false
}

//generate securely random URL safe token
func generateToken() string{
    b := make([]byte,32)
    _,err :=  rand.Read(b)
    if err != nil{
        return err.Error()
    }
    return base64.URLEncoding.EncodeToString(b)

}

func listenRPC() (net.Listener,error){
    rpc.Register(new(Communicator))
    rpc.HandleHTTP()
    return net.Listen("tcp",":"+strconv.Itoa(appConf.RPCPort))
}


func main() {
    f, ferr := os.OpenFile("/etc/zist/error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if ferr != nil {
		if os.IsPermission(ferr){
            fmt.Println("[*]","Can't create log file.Not enough priviledges.")
            return
        }
        fmt.Println(err.Error())
        return
	}
	defer f.Close()
    log.SetOutput(f)
    if parseCommand(){
        return
    }
    
	// confPath := flag.String("conf", "", "-conf=/path/to/conffile")
	keyfile := flag.String("keyfile", "", "-keyfile=path/to/keyfile")
	certfile := flag.String("certfile", "", "-certfile=path/to/certfile")
	flag.Parse()

	if err = ZistConf(); err != nil{
        fmt.Println("[*] Fatal:",err.Error())
        return
    }
    if err = ReadConfig(); err != nil{
        return
    }
    
	activeProcesses = make(map[int]*ChildProcess)

	for _, job := range jobs {
		go RegisterProcess(job, 0)
	}
    
    
    listener,rpcErr := listenRPC()
    if rpcErr != nil{
        log.Println(err.Error())
        return
    }
    
    //Start RPC listener
    go func(){
      log.Println(http.Serve(listener,nil).Error())  
    }()
    
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
	switch appConf.Protocol {
	case "http":
		http.ListenAndServe(":"+strconv.Itoa(appConf.HTTPPort), router)
	case "https": //key file and certificate file needed
		http.ListenAndServeTLS(":"+strconv.Itoa(appConf.HTTPPort), *certfile, *keyfile, router)
	default:
		log.Println("Invalid protocol:", appConf.Protocol)
	}
}
