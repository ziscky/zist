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
    "fmt"
    "io/ioutil"
    "path"
    "bufio"
    "errors"
    "github.com/BurntSushi/toml"
)

//ZistConfig stores the zistd configuration info
type ZistConfig struct{
    Confdir string
    Web bool
    Protocol string
    HTTPPort int
    RPCPort int
    Token string
}

var appConf ZistConfig

//Job represents the structure of monitored processes
type Job struct{
    Name string
    Path string
    Args string
    Workingdir string
    Logfile string
    Web bool
    Restart bool
}


var jobs []Job


//BinaryConf reads the binary config file .conf to find out the INSTALL_DIR and BINARY_DIR
func BinaryConf() error{
    _,err := os.Stat(".conf")
    if err != nil{
        if os.IsNotExist(err){
            fmt.Println("[*] Can't find the config installation path.")
            readConfigInput(1)
            createBinaryConf(1)
            return nil
        }
        return err
    }
    f,err1 := os.OpenFile(".conf",os.O_RDONLY,0777)
    if err1 != nil{
        return err1
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    scanner.Scan()
    INSTALL_DIR = scanner.Text()
    scanner.Scan()
    BINARY_DIR = scanner.Text()
    return nil
}
//ZistConf Reads zist configuration options
//stored by default in /etc/zist
func ZistConf() error{
       
       _,err := os.Stat(path.Join(INSTALL_DIR,"conf.toml"))
       if err != nil{
           if os.IsPermission(err){
               log.Println("Run zist with proper permissions")
               return err
           }
           
           if os.IsNotExist(err){
                log.Println("Can't find", INSTALL_DIR,"/conf.toml run `./zist install`")
               return err
           }
           return err           
       }
       
       if _,err1 := toml.DecodeFile(path.Join(INSTALL_DIR,"conf.toml"),&appConf); err1 != nil{
           log.Println(err1)
           return err1
       }
       if appConf.Token == ""{
           return errors.New("[*] Token cant be empty. Run `sudo zistd generate` to create a secure token. Add it to" + INSTALL_DIR + "/conf.toml")
       }
       if appConf.RPCPort == 0{
           return errors.New("[*] RPC port needed")
       }
       return nil
}

//ReadConfig reads and parses all config files 
func ReadConfig() error{
    info,err := os.Stat(appConf.Confdir)
    if err != nil{
        if os.IsPermission(err){
               log.Println("Run zist with proper permissions")
               return err
           }
           
           if os.IsNotExist(err){
               log.Println("Can't find", INSTALL_DIR + "/conf.d/ run `./zist install`")
               return err
           }
           return err    
    }
    if !info.IsDir(){
         log.Println("Can't find", INSTALL_DIR + "/conf.d/ run `./zist install`")
        return err
    }
    dir,err1 := ioutil.ReadDir(appConf.Confdir)
    if err1 != nil{
        log.Println(err1)
        return err1
    }
    if len(dir) < 1{
        fmt.Println("Nothing to run...Exiting")
        os.Exit(0)
    }
    for _,f := range dir{
        if err := ParseConfig(f); err != nil{
            log.Println("Error parsing " + f.Name())
            log.Println(err)
            continue
        } 
    }
    return nil  
}

//ParseConfig decodes job toml and adds to local []Job
func ParseConfig(f os.FileInfo) error{
    var job Job
    if _,err := toml.DecodeFile(path.Join(appConf.Confdir,f.Name()),&job); err != nil{
        return err
    }
    jobs = append(jobs,job)
    return nil
}