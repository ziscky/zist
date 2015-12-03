#**zist
A painkiller to use when deploying binaries (especially Go code) to manage/deploy instances of processes. Also carries a set of libraries to ease writing manageable Go code.

##The supervisor can:

-Start/Kill a process

-View process stats i.e cpu/mem usage

-Restart a process automatically

-Provide a stream of stdout/stderr output to terminal/twitter/browser

-Email/tweet/*text error logs to dev/sysadmin depending on the custom severity level

-Provide reverse proxying capabilities

##Ships with a fast deploy tool 
>For Go code which builds,tests (reporting failures) and deploys code automatically.

##Useful middleware and libraries:
-Commonly used header inclusion e.g CORS

-Profiling Request Handlers

-Basic customizable HTTP Auth

-Request Variable Storage //Only if you use the library it includes

-logging of errors to stdout/stderr/file depending on severity level

##PLANS
-Github and Jira integrations
