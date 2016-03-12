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
    "fmt"
    "os"
    "os/exec"
    "log"
    "errors"
)

//install 
//makes the conf dir 
//creates default conf.toml
//creates conf.d/
//TODO: move binaries
func install() error{
    if  _,err := os.Stat("/etc/zist"); err != nil{
        if os.IsPermission(err){
            log.Println("Run with correct permissions")
            return err
        }
        
        if os.IsNotExist(err){
              if err := makeDir(); err != nil{
                  log.Println(err)
                  return err
              }
        }else{
            log.Println(err)
            return err
        }
    }
    if _, err := os.Stat("/etc/zist/conf.toml"); err != nil{
        if os.IsNotExist(err){
            if err := createDefaultConfig(); err != nil{
                log.Println(err)
                return err
            }
        }else{
            log.Println(err)
            return err
        }
    }
    
    if err := createConfDir(); err != nil{
        log.Println(err)
        return err
    }
    if err := copyZistBinaries(); err != nil{
        return err
    }
    return nil
}
func makeDir() error{
    return os.MkdirAll("/etc/zist",0777)
}

func createConfDir() error{
    return os.MkdirAll("/etc/zist/conf.d",0777)
}

func createDefaultConfig() error{
    if  f,err := os.OpenFile("/etc/zist/conf.toml",os.O_CREATE | os.O_RDWR,0777); err != nil{
        return err
    }else{
        if _,err := f.WriteString("Confdir = \"/etc/zist/conf.d\"\nWeb = true\nProtocol = \"http\"\nHTTPPort = 7000\nRPCPort = 9876\nToken = "); err != nil{
         return err
        }
    }
    
    return nil
}

//copyZistBinaries copies the zistcl and zistd binaries to the /usr/bin dir
func copyZistBinaries() error{
    info,err := os.Stat("bin/")
    if err != nil{
        if os.IsNotExist(err){
            fmt.Println("[*] The bin/ dir does not exist. Run ./build.sh to install")
            return err
        }
        if os.IsPermission(err){
            fmt.Println("[*] Not enough priviledges.")
            return err
        }
        fmt.Println(err.Error())
        return err
    }
    if !info.IsDir(){
        fmt.Println("[*] The bin/ dir does not exist. Run ./build.sh to install")
        return errors.New("")
    }
    

    if _,err := os.Stat("bin/zistcl"); err != nil{
        if os.IsNotExist(err){
            fmt.Println("[*] Run ./build.sh")
            return err
        }
        fmt.Println(err.Error())
        return err
    }
    if err := exec.Command("cp","bin/zistcl","/usr/bin/zistcl").Run(); err != nil{
        fmt.Println("[*] Ensure you have suitable priviledges in /usr/bin")
        return err
    }
    
    if _,err := os.Stat("bin/zistd"); err != nil{
        if os.IsNotExist(err){
            fmt.Println("[*] Run ./build.sh")
            return err
        }
        fmt.Println(err.Error())
        return err
        
    }
    if  err := exec.Command("cp","bin/zistd","/usr/bin/zistd").Run(); err != nil{
        fmt.Println("[*] Ensure you have suitable priviledges in /usr/bin")
        return err
    }
    fmt.Println("[*] Succesfully installed the zist binaries.")
    return nil
}
