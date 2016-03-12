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
    "io/ioutil"
    "path"
    "errors"
    "github.com/BurntSushi/toml"
)

type ZistConfig struct{
    Confdir string
    Web bool
    Protocol string
    HTTPPort int
    RPCPort int
    Token string
}

var appConf ZistConfig

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

//ZistConf Reads zist configuration options
//stored by default in /etc/zist
func ZistConf() error{
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
       if appConf.Token == ""{
           return errors.New("[*] Token cant be empty. Run `sudo zistd generate` to create a secure token. Edit /etc/zist/conf.toml")
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
               log.Println("Can't find /etc/zist/conf.d/ run `sudo zist install`")
               return err
           }
           return err    
    }
    if !info.IsDir(){
        log.Println("Can't find /etc/zist/conf.d/ run `sudo zist install`")
        return err
    }
    dir,err1 := ioutil.ReadDir(appConf.Confdir)
    if err1 != nil{
        log.Println(err1)
        return err1
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