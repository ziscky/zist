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
    "net/rpc"
    "os"
    "log"
    "strconv"
    "fmt"
    "github.com/BurntSushi/toml"
)



//Purpose of the cli tool is to interface with the running zistd process.
//-Status of zistd
//-Status of zistd monitored processes
//-Reload zistd monitored process configs
//-Kill zistd

//ZistConfig represents the zistd config structure
type ZistConfig struct{
    Confdir string
    Web bool
    Protocol string
    HTTPPort int
    RPCPort int
    Token string
}

var appConf ZistConfig

//Kill instructs zistd to terminate
func Kill(client *rpc.Client,flag int) (string,error){
    var status string
    return status,client.Call("Communicator.Kill",flag,&status)
}

//Reload causes zistd to reload the monitored process configs
func Reload(client *rpc.Client) (string,error){
    var status string
    return status,client.Call("Communicator.Reload",0,&status)
}

//DStatus inquires the status of zistd
func DStatus(client *rpc.Client) (string,error){
    var status bool
    if err := client.Call("Communicator.Status",0,&status); err !=nil{
        return "Error",err
    }
    if status{
        return "zistd is alive!",nil
    }
    return "zistd is not running",nil
}

//VerifyToken authenticates the zistcl request to zistd
func VerifyToken(client *rpc.Client,token string) (bool,error){
    var valid bool
    if err := client.Call("Communicator.VerifyToken",token,&valid); err != nil{
        return false,err
    }
    return valid,nil
}

//ProcStatus gets the process info of a particular process 
//gets back a json string
func ProcStatus(client *rpc.Client,pname string) (string,error){
    var status string
    return status,client.Call("Communicator.ProcessStatus",pname,&status)
}

//ProcStart starts a  monitored process by name
func ProcStart(client *rpc.Client,pname string) (string,error){
    var status string
    return status,client.Call("Communicator.ProcessStart",pname,&status)
}

//ProcRestart restarts a monitored process by name
func ProcRestart(client *rpc.Client,pname string) (string,error){
    var status string
    return status,client.Call("Communicator.ProcessRestart",pname,&status)
}

//ProcDetach instructs zistd to detach a monitored process
func ProcDetach(client *rpc.Client,pname string) (string,error){
    var status string
    return status,client.Call("Communicator.ProcessDetach",pname,&status)
}

//ProcStop instructs zistd to stop a monitored process by name
func ProcStop(client *rpc.Client,pname string) (string,error){
    var status string
    return status,client.Call("Communicator.ProcessStop",pname,&status)
}

//GetLog gets the contents of zistd log file
func GetLog(client *rpc.Client) (string,error){
     var status string
     return status,client.Call("Communicator.ReadLog",0,&status)
}

//ClearLog clears the contents of the zistd log file
func ClearLog(client *rpc.Client) (string,error){
    var status string
    return status,client.Call("Communicator.ClearLog",0,&status)
}

 var arg1,arg2,arg3,arg4 string

//setLocal assigns arguments if -l switch is present
func setLocal() bool{
    switch len(os.Args) {
        case 2:
            arg1 = os.Args[1]
            break
        case 3:
            arg1 = os.Args[1]
            arg2 = ""
            arg3 = os.Args[2]
            break
        case 4:
            arg1 = os.Args[1]
            arg2 = ""
            arg3 = os.Args[2]
            arg4 = os.Args[3]
            break
         default:
            printUsage()
            return false
        }
        return true
}

//setRemote assigns arguments assuming connection to remote zistd instance
func setRemote() bool{
    switch len(os.Args) {
    case 2:
        arg1 = os.Args[1]
        break
    case 3:
        arg1 = os.Args[1]
        arg2 = os.Args[2]
        break
    case 4:
        arg1 = os.Args[1]
        arg2 = os.Args[2]
        arg3 = os.Args[3]
        break
    case 5:
        arg1 = os.Args[1]
        arg2 = os.Args[2]
        arg3 = os.Args[3]
        arg4 = os.Args[4]
        break
    default:
        printUsage()
        return false
    }
    return true
}

func main() {
    //connect to the rpc server: default localhost
    //zist http:1.1.1.1:9000 ___________ cmds
    
    //set arguments
    if local(){
        if !setLocal(){
            return
        }
    }else{
        if !setRemote(){
            return
        }
    }
    
    if arg1 == "-h"{
        printUsage()
        return
    }
    
    var client *rpc.Client
    var err error
    
    //connect to rpc server
    if arg1 == "-l"{
        if err := readLocalConfig(); err != nil{
            return
        }
        client,err = rpc.DialHTTP("tcp",":"+strconv.Itoa(appConf.RPCPort))
    }else{
        client,err = rpc.DialHTTP("tcp",os.Args[1])
    }        
   
    if err != nil{
        fmt.Println(err)
        return
    }
    defer client.Close()
    
    //verify token
    if !local(){
        if verified,err := VerifyToken(client,arg2); err != nil{
            log.Println(err)
        }else{
            if !verified{
                fmt.Println("[*] Invalid RPC access token.")
                return
            }
        }
    }
    
    
    if arg3 == ""{
        printUsage()
        return
    }
    //handle status    
    if arg3 == "status"{
        stat,err1 := DStatus(client)
        if err1 != nil{
            fmt.Println(err1.Error())
            return
        }
        fmt.Println("[*]",stat)
        return
    }
    
    //handle logs
    if arg3 == "log"{
        if arg4 == ""{
            stat,err := GetLog(client)
            if err != nil{
                fmt.Println(err.Error())
                return
            }
            fmt.Println("[*]",stat)
        }else if arg4 == "clear"{
            stat,err := ClearLog(client)
            if err != nil{
                fmt.Println(err.Error())
                return
            }
            fmt.Println("[*]",stat)
        }else{
            fmt.Println("[*] Unknown command",arg4)
            printUsage()
            return
        }
        return
    }
    //handle kill
    if arg3 == "kill"{
        if arg4 == ""{
            fmt.Println("[*] Retaining monitored processes.Killing zistd.")
            Kill(client,1)
            return
        }
        if arg4 == "all"{
            fmt.Println("[*] Killing all processes together with zistd.")
            Kill(client,0)    
        }else{
            fmt.Println("[*] Unknown command " + arg4)
        }
        return
    }
    
    //handle reload
    if arg3 == "reload"{
        stat,err := Reload(client)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
    }
    
    
    //handle childprocess inquiries
    if arg4 == ""{
        fmt.Println("[*] Invalid usage")
        printUsage()
        return
    }
    
    switch arg4 {
    case "status":
        stat,err := ProcStatus(client,arg3)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
    case "start":
        stat,err := ProcStart(client,arg3)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
    case "stop":
        stat,err := ProcStop(client,arg3)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
     case "restart":
        stat,err := ProcRestart(client,arg3)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
      case "detach":
        stat,err := ProcDetach(client,arg3)
        if err != nil{
            log.Println(err)
            return
        }
        fmt.Println("[*]",stat)
        return
       default:
        fmt.Println("[*] Unknown command",arg4)
        return
    }
    
}

//local checks if long command or short command
func local() bool{
    if len(os.Args) < 2{
        return false
    }
    if os.Args[1] == "-l"{
        return true
    }
    return false
}

//readLocalConfig reads the local /etc/zist/conf.toml and parses to appConf var.
func readLocalConfig() error{
         _,err := os.Stat("/etc/zist/conf.toml")
       if err != nil{
           if os.IsPermission(err){
               log.Println("Run zist with proper permissions")
               return err
           }
           
           if os.IsNotExist(err){
               log.Println("Can't find /etc/zist/conf.toml run `sudo zist install`")
               return err
           }
           return err           
       }
       
       if _,err1 := toml.DecodeFile("/etc/zist/conf.toml",&appConf); err1 != nil{
           log.Println(err1)
           return err1
       }
       return nil
}


func printUsage(){
    fmt.Println(`Usage: zistcl host:port token [cmd] or zistcl -l [cmd] #for local connection
        CMDS: kill - kill zistd but detach monitored procs to continue running on their own
              kill all - kill zistd and monitored procs
              status - get status of zistd
              reload - reload zistd monitored process config files
              log - get the contents of the zistd log file
              log clear - clear the contents of the zistd log file
              example: zistcl 1.1.1.1:2000 kill all
              
        PROCESS CMDS: i.e zistcl host:port -l [processname] [cmd] or zistcl -l [processname] [cmd]
              status - get status of the named process
              restart - restart the named process
              stop - stop the named process
              detach - detach the named process from zistd
              start - start the named process
              example: zistcl -l app1 restart`)
   
}
