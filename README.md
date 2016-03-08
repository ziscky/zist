# **zist**
[![Build Status](https://travis-ci.org/ziscky/zist.svg?branch=master)](https://travis-ci.org/ziscky/zist)

#Think supervisord plus remote commandeering via rpc
------------------------------------------------------------------------

##Supervisor functions
Register processes to be started and monitored in a familiar way using an easy to write toml file.
Sample:
Name = "app1"
Path = "/path/to/app1"
Args = "-your -app -args"
Restart = true
Web = true


## Live Process Interaction via web API and RPC client
 - View process stats i.e cpu/mem usage
 - Start/Stop/Restart a process
 - Detach a process to allow it to run without zist supervision
 - View a list of all managed processes and valuable information e.g
  *[{"isalive":true,"name":"xm","numrestarts":1,"path":"/home/eziscky/Golang/src/github.com/ziscky/tests/xm","pid":2067,"timealive":"13.410013177s","timestarted":"2015-12-17 13:36:44.540403109 +0300 EAT"}]*
 - Get stdout/stderr process logs and log files.

##Process Detach
 - If you don't like the *coupling of your proc's instance with zist* you
   can simply detach it after you're done monitoring it
**N.B** *Process Detach is only supported for direct invocation of binaries. So if you run a sh script, I'm working to support that. Or just send in a pull request :)*

#FUTURE
 - Live process hijacking.
 
 #USAGE
 ---------------------------------------------------------------------------------------------
 
##Installation

###Building from source
You will need to install Go (*https://golang.org*)
You can clone this repo: *https://github.com/ziscky/zist* or use go get *https://github.com/ziscky/zist*
In the zist directory run the bash script *build.sh*
Run *sudo zist install*, or not *sudo* if you have correct permissions for:
        - /usr/bin
        - /etc 
On installation zistd and zistcl will be installed in */usr/bin*. Typically for remote administration all you need on the server is zistd.

###Configuration
Configurations are stored in */etc/zist*

The file of interest is */etc/zist/conf.toml*

It typically looks like this:
    - Maintain quotes for strings.
           
    Confdir = "/etc/zist/conf.d"
    Web = true
    Protocol = "http"
    HTTPPort = 7000
    Token = "zis"
    RPCPort = 9876

These are the defaults.

    - Confdir: where zistd will look for process config files
    - Web: if zistd will render process information via the web API
    - Protocol: http or https
    - HTTPPort: where the web API server will listen
    - RPCPort: where the rpc server will listen
    - Token: used to secure API/RPC connections

A few gotchas:
    - Strings in the config file should be in quotation. View *https://github.com/toml-lang/toml* for more on toml.
    - For https you have to provide the keyfile and certfile when running zist
    
Put your app config files in the Confdir directory specified in *conf.toml*
    - Each app/process/script must have its own file
    - The name of the file does not matter
    
    Name = "name of the process.(no spaces or special chars)"
    Path = "/path/to/process"
    Args = "-your -app -args"
    Restart = true/false
    Web = true/false 
    

###Running
Run it with *nohup zistd &* to run it in the background.
To generate a secure token:
    - Run *zistd generate*
    - Copy the token to your config file

###Interaction

There are 2 ways to interact with a running zistd instances.
####1. zistcl
       zistcl is a commandline tool to interact with zistd.(Kinda like supervisord supervisorctl relationship)
       *You can get the following info and more by running zistcl -h*
       To connect to a zistd instance:
            *zistcl host:port token [cmd]* for remote connection to zistd
            *zistcl -l [cmd]* for local connection
            
            Examples:
            zistcl 1.1.1.1:9876 mysecuretoken status //Get zistd status
            zistcl -l log //get zistd log output
            zistcl -l app1 status //get app1 status
            zistcl 1.1.1.1:9876 mysecuretoken  app1 detach


####2. Web API
        Here are the API routes:
        Output is standard JSON.
                host:port/{token} -> Gets All the monitored process info, including pid that can be used in the below requests 
                host:port/{token}/{pid}/stats
                host:port/{token}/{pid}/kill 
                host:port/{token}/{pid}/start
                host:port/{token}/{pid}/restart
                host:port/{token}/{pid}/stdout
                host:port/{token}/{pid}/stderr
                host:port/{token}/{pid}/detach

#NOTE
    - Beta software do not use in prod
    - Feel free to contribute
    - Tell me what you do with it