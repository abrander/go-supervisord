# go-supervisord
RPC remote control for [supervisord](http://supervisord.org/)

[![GoDoc][1]][2]

[1]: https://godoc.org/github.com/abrander/go-supervisord?status.svg
[2]: https://godoc.org/github.com/abrander/go-supervisord

Code Examples
-------------

Reloading configuration and clearing daemon log:
```go
import "github.com/abrander/go-supervisord"
  
func main() {
	c, err := supervisord.NewClient("http://127.0.0.1:9001/RPC2")
	if err != nil {
		panic(err.Error())
	}
	
	err = c.ClearLog()
	if err != nil {
		panic(err.Error())
	}
	
	err = c.Restart()
	if err != nil {
		panic(err.Error())
	}
}
```

Stop supervisord process `worker`:
```go
import "github.com/abrander/go-supervisord"
  
func main() {
	c, err := supervisord.NewClient("http://127.0.0.1:9001/RPC2")
	if err != nil {
		panic(err.Error())
	}
	
	err = c.StopProcess("worker", false)
	if err != nil {
		panic(err.Error())
	}
}
```
