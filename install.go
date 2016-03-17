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
    "path/filepath"
    "log"
    "path"
    "errors"
    "bufio"
)

//install 
//makes the conf dir 
//creates default conf.toml
//creates conf.d/


var(
    //INSTALL_DIR is the location of the zist config install
     INSTALL_DIR = "/etc/zist"
     //BINARY_DIR is the location of the zist bins
     BINARY_DIR = "/usr/bin"
)

func readConfigInput(configonly int){
    scanner := bufio.NewScanner(os.Stdin)
    
        fmt.Print("Config Installation Path (leave blank for default: /etc/zist ): ")
        scanner.Scan()
        text := scanner.Text()
        if len(text) > 0{
            INSTALL_DIR = text
        }
        if configonly > 0{
            return
        }
    
    fmt.Print("Binary Installation Path (leave blank for default: /usr/bin ): ")
    scanner.Scan()
    text2 := scanner.Text()
    if len(text2) > 0{
        BINARY_DIR = text2
    }
}
func install() error{
    //ask user for install dir
    readConfigInput(0)
    if  _,err := os.Stat(INSTALL_DIR); err != nil{
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
    if _, err := os.Stat(path.Join(INSTALL_DIR,"conf.toml")); err != nil{
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
    if err := createLogFile();err != nil{
        log.Println(err)
        return err
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
    return os.MkdirAll(INSTALL_DIR,0777)
}

func createConfDir() error{
    return os.MkdirAll(path.Join(INSTALL_DIR,"conf.d"),0777)
}

func createDefaultConfig() error{
    if  f,err := os.OpenFile(path.Join(INSTALL_DIR,"conf.toml"),os.O_CREATE | os.O_RDWR,0777); err != nil{
        return err
    }else{
        if _,err := f.WriteString("Confdir = \""+ path.Join(INSTALL_DIR + "/conf.d") +"\"\nWeb = true\nProtocol = \"http\"\nHTTPPort = 7000\nRPCPort = 9876\nToken = \"changeme\""); err != nil{
         return err
        }
    }
    
    return nil
}

func createLogFile() error{
    if _,err := os.OpenFile(path.Join(INSTALL_DIR,"error.log"),os.O_CREATE | os.O_RDWR,0777); err != nil{
       return err 
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
    
     if err := createBinaryConf(0); err != nil{
         return err
     }
    
    if err := exec.Command("cp","bin/zistcl",path.Join(BINARY_DIR,"zistcl")).Run(); err != nil{
        fmt.Println("[*] Ensure you have suitable priviledges in",BINARY_DIR)
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
    if  err := exec.Command("cp","bin/zistd",path.Join(BINARY_DIR,"zistd")).Run(); err != nil{
        fmt.Println("[*] Ensure you have suitable priviledges in",BINARY_DIR)
        return err
    }

    fmt.Println("[*] Succesfully installed the zist binaries.")
    return nil
}

func createBinaryConf(wdir int) error{
    var f *os.File
    var err  error
    dir,_ := filepath.Abs(filepath.Dir(os.Args[0]))//current binary directory
    if wdir > 0{
        f,err = os.OpenFile(path.Join(dir,".conf"),os.O_CREATE | os.O_RDWR,0777)
    }else{
        f,err = os.OpenFile(path.Join(BINARY_DIR,".conf"),os.O_CREATE | os.O_RDWR,0777)
    }
    if err != nil{
        if wdir > 0{
            fmt.Println("[*] Unable to move configuration data to",dir)
        }else{
            fmt.Println("[*] Unable to move configuration data to",BINARY_DIR)
        }
        return err
    }
    defer f.Close()
    if wdir > 0{
        f.WriteString(INSTALL_DIR + "\n" + dir)
    }else{
        f.WriteString(INSTALL_DIR + "\n" + BINARY_DIR)
    }
    return nil
}