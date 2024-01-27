package geerpc

import (
	"fmt"
	"net/http"
	"text/template"
)

// debugText不需要关注过多
const debugText = `<html>
	<body>
	<title>GeeRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugHTTP struct {
	*Server //继承做法
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

// Runs at /debug/rpc， 调用的是debugHTTP的ServeHTTP，不是server结构体的ServeHTTP
func (server debugHTTP) ServerHTTP(w http.ResponseWriter, rep *http.Request) {
	var services []debugService
	//sync.Map遍历,Range方法并配合一个回调函数进行遍历操作。通过回调函数返回遍历出来的键值对。
	server.serviceMap.Range(func(namei, svci any) bool {
		svc := svci.(*service) //转换成*service类型
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svc.method,
		})
		return true //当需要继续迭代遍历时，Range参数中回调函数返回true;否则返回false
	})

	err := debug.Execute(w, services)
	if err != nil {
		fmt.Fprintln(w, "rpc: error executing template:", err.Error())
	}
}
