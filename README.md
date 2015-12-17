# **zist**

#A simple modern server program to manage and interact process instances.
------------------------------------------------------------------------

##Start Processes (with automatic process restarting)
Start processes using an easy to write config file. Check the examples directory.All you need is:

 - Port
 -  Protocol ( *For https a keyfile and certfile are required obviously*)
 -  Access Token
 - Process Path
 - Process Name

That's All!!

## Live Process Interaction
 - View process stats i.e cpu/mem usage
 -  Start a process and Kill it
 - Restart a process
 - View a list of all managed processes and valuable information e.g
 - *[{"isalive":true,"name":"xm","numrestarts":1,"path":"/home/eziscky/Golang/src/github.com/ziscky/tests/xm","pid":2067,"timealive":"13.410013177s","timestarted":"2015-12-17 13:36:44.540403109 +0300 EAT"}]*
 - Stream stdout/stderr output to a web endpoint, with pretty basic JSON output

##Crash Reporting
 1. Send a mail of the managed instance context(*configurable*) after the
    crash to a list of emails.

##Process Detach
 - If you don't like the *coupling of your proc's instance with zist* you
   can simply detach it after you're done monitoring it!!
**N.B** *Process Detach is only supported for direct invocation of binaries. So if you run a sh script, I'm working to support that. Or just send in a pull request :)*

##Process Injection
 - Support for config injection to add new processes without restarting zist is in dev. *Basically Live Config Reloading.*

#NOTE
 - Still under moderately heavy development so expect some changes
 - Still undergoing refactoring
 - Mail Logging is not complete but other supervisor functions are OK.
