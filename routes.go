package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var memstore = make(map[*http.Request]map[string]interface{})
var memutex sync.Mutex

func StoreVar(r *http.Request, key string, v interface{}) {
	memutex.Lock()
	defer memutex.Unlock()
	if memstore[r] == nil {
		memstore[r] = make(map[string]interface{})
	}
	memstore[r][key] = v
}

func RemoveVars(r *http.Request) {
	memutex.Lock()
	defer memutex.Unlock()
	delete(memstore, r)
}

func GetVar(r *http.Request, key string) interface{} {
	memutex.Lock()
	defer memutex.Unlock()
	return memstore[r][key]
}

func CheckToken(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars["token"] != appConf.Ep.Token {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(rw, r)
	}
}

func WithProcess(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pid, err := strconv.Atoi(vars["pid"])
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		proc, exists := activeProcesses[pid]
		if !exists {
			http.Error(rw, "Process Does Not Exist", http.StatusNotFound)
			return
		}
		StoreVar(r, "proc", proc)
		StoreVar(r, "pid", pid)
		next(rw, r)
	}
}

func Stats(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	stats, sterr := proc.Stats()
	if sterr != nil {
		http.Error(rw, sterr.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(stats)
}

func Default(rw http.ResponseWriter, r *http.Request) {
	procs := []map[string]interface{}{}
	for _, proc := range activeProcesses {
		procs = append(procs, map[string]interface{}{
			"pid":         proc.PID,
			"name":        proc.Pname,
			"path":        proc.PPath,
			"numrestarts": proc.RestartCount,
			"timestarted": proc.Timestamp.String(),
			"timealive":   time.Since(proc.Timestamp).String(),
			"isalive":     proc.IsAlive,
		})
	}
	json.NewEncoder(rw).Encode(procs)
}

func Kill(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	proc.Kill()
	log.Println(proc.PID, " killed")
	rw.Write([]byte(strconv.Itoa(proc.PID) + " killed"))
}

func Start(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	pid := GetVar(r, "pid").(int)
	defer RemoveVars(r)

	if err := proc.Initialize(proc.PPath); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	AddProcess(proc)

	cp := new(ChildProcess)
	cp.PID = pid
	RemoveProcess(cp)

	log.Println(pid, " restarted with new pid ", proc.PID)
	rw.Write([]byte(strconv.Itoa(proc.PID)))
}

func Restart(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	pid := GetVar(r, "pid").(int)
	defer RemoveVars(r)

	proc.Kill()

	if err := proc.Initialize(proc.PPath); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	AddProcess(proc)

	cp := new(ChildProcess)
	cp.PID = pid
	RemoveProcess(cp)

	log.Println(pid, " restarted with new pid ", proc.PID)
	rw.Write([]byte(strconv.Itoa(proc.PID)))
}

func Detach(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	if err := proc.Detach(); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	rw.Write([]byte(proc.Pname + " has been successfully detached. I will no longer restart it if it fails,give you stdstreams  or give stats"))
}

func StdOut(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	json.NewEncoder(rw).Encode(proc.GetOutput())
}

func StdErr(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	json.NewEncoder(rw).Encode(proc.GetErrors())
}
