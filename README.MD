# Goze

Powerful Golang `Web Framework`

## Features

* RESTful HTTP Server
* Dependence injection
* Configurations (YAML)
* Auto SQL Transaction


## Installation

Install via go get.

``` shell
$ go get github.com/azzill/goze
```

## Example

Start a server on :8080 with a configured string

### Custom Configuration
main.go
```go
package main

import (
	"github.com/azzill/goze/bootstrap"
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/server"
	"github.com/azzill/goze/config"
)



type Controller struct{
	SQL *SQL `inject:"true"`
}

func (c *Controller) Mapping(s *server.RestServer) {
	s.GET("/", func(ctx *common.RequestCtx) (i interface{}) {
		return c.SQL.url
	})

}

type SQL struct {
	url string
}

func (s *SQL) Config(cfg *config.CommonConfiguration) interface{} {
	s.url = cfg.Get("sql.url").(string)
	return s
}


func main() {
		bootstrap.StartGozeApplication(&Controller{},&SQL{})
}

```
server.yaml
```yaml
sql:
  url: localhost
```

### Interceptor
```go

package demo

import (
	"fmt"
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/midware"
)

type CustomInterceptor struct {

}

func (*CustomInterceptor) Priority() int {
	return 0
}

func (*CustomInterceptor) Intercept(ctx *common.RequestCtx) (midware.InterceptorAction, interface{}) {
	fmt.Println("Interceptor call!")
	return midware.Continue, nil
}

```

### ResponseWrapper
```go

package demo

import (
	"fmt"
	"net/http"
)

type CustomWrapper struct {

}

func (*CustomWrapper) Wrap(v interface{}, wr http.ResponseWriter) bool {
	switch v.(type) {
	case string:
		_, _ = fmt.Fprintln(wr, "Wrapped:", v)
		return true
	}
	return false
}


```

## Dependencies

* [garyburd/redigo v1.6.0](https://github.com/garyburd/redigo)
* [go-yaml/yaml v2.2.2](https://github.com/go-yaml/yaml)