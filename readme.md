## DockerMonitor
监控docker状态，日志关键字报警，钉钉报警，设置静默时间

docker status, log alert,Rev dingtalk alert,set silence duration
## Usage

./dockermonitor [OPTION...]
```
-f  how many mins for silence (default: "1")

-k  keyworks for identify split by ',' 
(default: "retrying,abort")

-p  prefix (default: "l2")

-y  locate db.yaml (default: "./db.yaml")

-h  show help (default: false)
```

## Config
**main.go**
```go

const (
	dingToken    = "--your dingToken--"
	secretString = "--your secret--"
)
```

## Example
**Keyworkds '*retrying,abort*' expected as an error and set 60 mins silence time for a same error**
```
./dockermonitor -f 60 -k retrying,abort -p l4 -y /root/docker_monitor_exe/db.yaml > /root/docker_monitor_exe/dm.log 2>&1
```
**Recieve alert msg from dingtalk**
```
------start l4 2021-06-30 13:26:01------
/hd12-tmp1 error 
/hd12-tmp2 error 
------end------
```

