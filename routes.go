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
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var memstore = make(map[*http.Request]map[string]interface{})
var memutex sync.Mutex

//StoreVar stores a request var in the request map
func StoreVar(r *http.Request, key string, v interface{}) {
	memutex.Lock()
	defer memutex.Unlock()
	if memstore[r] == nil {
		memstore[r] = make(map[string]interface{})
	}
	memstore[r][key] = v
}

//RemoveVars removes the request vars from the request map
func RemoveVars(r *http.Request) {
	memutex.Lock()
	defer memutex.Unlock()
	delete(memstore, r)
}

//GetVar gets a variable from the request map
func GetVar(r *http.Request, key string) interface{} {
	memutex.Lock()
	defer memutex.Unlock()
	return memstore[r][key]
}

//CheckToken checks the request token
func CheckToken(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars["token"] != appConf.Token {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(rw, r)
	}
}

//WithProcess adds the process to the request var
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

//Stats returns the process stats via web API
func Stats(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	if !proc.EStats {
		rw.Write([]byte("Not allowed"))
		return
	}
	stats, sterr := proc.Stats()
	if sterr != nil {
		http.Error(rw, sterr.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(rw).Encode(stats)
}

//Default is the default API route ,returns all monitored processs info
func Default(rw http.ResponseWriter, r *http.Request) {
	procs := []map[string]interface{}{}
	for _, proc := range activeProcesses {
		procs = append(procs, map[string]interface{}{
			"pid":         proc.PID,
			"name":        proc.Pname,
			"path":        proc.PPath,
			"args":        proc.Args,
			"numrestarts": proc.RestartCount,
			"timestarted": proc.Timestamp.String(),
			"timealive":   time.Since(proc.Timestamp).String(),
			"isalive":     proc.IsAlive,
		})
	}
	json.NewEncoder(rw).Encode(procs)
}

//Kill instructs zist to kill a process by pid
func Kill(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	proc.Kill()
	log.Println(proc.PID, " killed")
	rw.Write([]byte(strconv.Itoa(proc.PID) + " killed"))
}

//Start instructs zist to start a process by PID
func Start(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	pid := GetVar(r, "pid").(int)
	defer RemoveVars(r)

	if err := startProcess(proc, pid, 0); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	log.Println(pid, " restarted with new pid ", proc.PID)
	rw.Write([]byte(strconv.Itoa(proc.PID)))
}

//Restart instructs zist to restart a process
func Restart(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	pid := GetVar(r, "pid").(int)
	defer RemoveVars(r)

	proc.Kill()
	if err := startProcess(proc, pid, proc.RestartCount+1); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	log.Println(pid, " restarted with new pid ", proc.PID)
	rw.Write([]byte(strconv.Itoa(proc.PID)))
}

//Detach requests zist to detach a process
func Detach(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	if err := proc.Detach(); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	rw.Write([]byte(proc.Pname + " has been successfully detached. I will no longer restart it if it fails,give you stdstreams  or give stats"))
}

//StdOut gets the process stdout
func StdOut(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	if !proc.EStdOut {
		rw.Write([]byte("Not allowed"))
		return
	}
	json.NewEncoder(rw).Encode(proc.GetOutput())
}

//StdErr gets the process stderr
func StdErr(rw http.ResponseWriter, r *http.Request) {
	proc := GetVar(r, "proc").(*ChildProcess)
	defer RemoveVars(r)
	if !proc.EStdErr {
		rw.Write([]byte("Not allowed"))
		return
	}
	json.NewEncoder(rw).Encode(proc.GetErrors())
}

//startProcess starts the requested child process
func startProcess(proc *ChildProcess, pid, numrestarts int) error {
	procstrct := Job{
		Path: proc.PPath,
	}
	if err := proc.Initialize(procstrct, numrestarts); err != nil {
		return err
	}
	proc.IsAlive = true
	AddProcess(proc)

	cp := new(ChildProcess)
	cp.PID = pid
	RemoveProcess(cp)
	return nil
}
