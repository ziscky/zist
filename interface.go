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

import(
    "os"
    "log"
    "time"
    "bufio"
    "encoding/json"
)
//RPC handlers working as an interface to the cli tool

//Communicator handles all remote communication with zistd following the gorpc structure
type Communicator struct{}


//VerifyToken verifies the given token from zistcl
func (comm *Communicator) VerifyToken(token string,valid *bool) error{
    *valid = appConf.Token == token
    return nil
}
//Kill kills zistd
//i=0 kill with all monitored processes
//i=1 detach all processes then kill
func (comm *Communicator) Kill(i int,msg *string) error{
   for _,proc := range activeProcesses{
       if i == 0{
          if err := proc.Kill(); err != nil{
            *msg += proc.Pname + " failed to exit." + err.Error()
          }
          continue
       }
       if err := proc.Detach(); err != nil{
           *msg += proc.Pname + " failed to detach." + err.Error()
       }
   }
   os.Exit(0)
   return nil
}

//Reload reloads the monitored process configs
//Also reload zistd config??
func (comm *Communicator) Reload(_ int,msg *string) error{
   for _,proc := range activeProcesses{
       if err := proc.Kill(); err != nil{
           log.Println(err)
           continue
       }
       RemoveProcess(proc)
   }
   jobs = jobs[:0]
   if err := ReadConfig(); err != nil{
       *msg  = err.Error()
       return nil
   }
   
   for _,job := range jobs{
       go RegisterProcess(job,0)
   }
   *msg = "Succesfully reloaded configs"
   return nil
}

//Status checks if zistd is alive
func (comm *Communicator) Status(_ int, ack *bool) error{
    *ack = true
    return nil
}

//ProcessStatus gets the overall status of a process by name as defined
//in the proc config file
func (comm *Communicator) ProcessStatus(name string,msg *string) error{
   procLock.Lock()
   defer procLock.Unlock()
   for _,proc := range activeProcesses{
       if proc.Pname == name{
           payload := map[string]interface{}{
               "pid":         proc.PID,
			"name":        proc.Pname,
			"path":        proc.PPath,
			"args":        proc.Args,
			"numrestarts": proc.RestartCount,
			"timestarted": proc.Timestamp.String(),
			"timealive":   time.Since(proc.Timestamp).String(),
			"isalive":     proc.IsAlive,
           }
           payloadJson,err := json.Marshal(payload)
           if err != nil{
               log.Println(err)
               return err
           }
           *msg = string(payloadJson)
           return nil
       }
   }
   *msg = "No such process"
   return nil
}

//ProcessStop stops the requested process by name
func (comm *Communicator) ProcessStop(name string,msg *string) error{
   for _,proc := range activeProcesses{
       if proc.Pname == name{
           if err := proc.Kill(); err != nil{
               *msg = err.Error()
               return err
           }
           proc.IsAlive = false
           *msg = "Succesfully stopped"
           return nil
       }
   }
   *msg = "No such process"
   return nil
}

//ProcessDetach detaches the child process to become it's own process, losing state ofcourse
func (comm *Communicator) ProcessDetach(name string,msg *string) error{
   for _,proc := range activeProcesses{
       if proc.Pname == name{
           if err := proc.Detach(); err != nil{
               *msg = err.Error()
               return err
           }
           *msg = "Succesfully detached"
           return nil
       }
   }
   *msg = "No such process"
   return nil
}

//ProcessRestart restarts the process by name 
func (comm *Communicator) ProcessRestart(name string,msg *string) error{
   for _,proc := range activeProcesses{
       if proc.Pname == name{
           if err := proc.Kill(); err != nil{
               *msg = err.Error()
               return err
           }
           if err := startProcess(proc,proc.PID,proc.RestartCount+1); err != nil{
               *msg = err.Error()
               return err
           }
           *msg = "Succesfully restarted"
           return nil
       }
   }
   *msg = "No such process"
   return nil
}

//ProcessStart starts a monitored process by name
func (comm *Communicator) ProcessStart(name string,msg *string) error{
   for _,proc := range activeProcesses{
       if proc.Pname == name{
           if proc.IsAlive{
               *msg = "Process already running"
               return nil
           }
           if err := startProcess(proc,proc.PID,0); err != nil{
               *msg = err.Error()
               return err
           }
           *msg = "Succesfully started"
           return nil
       }
   }
   *msg = "No such process"
   return nil
}

//ReadLog reads zistd output log
func (comm *Communicator) ReadLog(_ int,msg *string) error{
    f, err := os.OpenFile("/etc/zist/error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		*msg = err.Error()
        return err
	}
	defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan(){
        *msg += "\n" + scanner.Text()
    }
    return nil
    
}

//ClearLog clears zistd output log
func (comm  *Communicator) ClearLog(_ int,msg *string) error{
    f, err := os.OpenFile("/etc/zist/error.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		*msg = err.Error()
        return err
	}
	defer f.Close()
    _,err1 := f.WriteString("")
    if err1 != nil{
        *msg = err1.Error()
        return err
    }
    *msg = "Succesfully cleared zistd log"
    return nil
}

