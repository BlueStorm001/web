package web1

import (
	"fmt"
	"testing"
	"time"
	proto3 "web/proto"
)

var ws = Service()

func init() {
	ws.NoHeader = true
}
func TestServe(t *testing.T) {

	ws.Timeout = time.Second * 60
	ws.Request = func(request proto3.Request) *proto3.Request {
		query := QueryForm(request.Path)
		fmt.Println(query.Path, query.Get["cid"], string(request.Body))
		//构造返回数据
		request.HttpStatusCode = "200"
		request.ContentType = "text/html"
		request.Body = []byte("hello world")
		request.Header = nil
		//异步返回
		ws.Response(request)
		//返回有内容直接Response
		return nil //&request
	}
	ws.StartRun("80")
}

func TestQueryForm(t *testing.T) {
	var request = proto3.Request{}
	request.Path = "/a/b/c/d/search.ashx?cid=888&b=wudongwen&host=r-=com&instanceId=r-uf6m0t1gzn04hfw7q3&regionId=cn"
	q := QueryForm(request.Path)
	fmt.Println(q)
}

func BenchmarkQueryForm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := QueryForm("/search.ashx?cid=888&b=wudongwen&host=r-=com&instanceId=r-uf6m0t1gzn04hfw7q3&regionId=cn")
		if q == nil {
			fmt.Println(q)
		}
	}
}

var body = []byte(`POST /api.ashx HTTP/1.1
Content-Type: application/json
Accept-Encoding: gzip
User-Agent: PostmanRuntime/7.29.0
Accept: */*
Postman-Token: 885a410f-f326-4326-958c-fa4cf2e99333
Host: 127.0.0.1:8086
Connection: keep-alive
Content-Length: 217

{
    "cid": "AIRPAZ",
    "tripType": "2",
    "fromCity": "HKG",
    "toCity": "LAX",
    "fromDate": "20220728",
    "retDate": "20220810",
    "adultNumber": 1,
    "childNumber": 0,
    "channel": "F"
}`)

func TestRequest(t *testing.T) {
	req := ws.httpRequest(body)
	fmt.Println(req)
}

func BenchmarkRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req := ws.httpRequest(body)
		if req.Message == "" {

		}
	}
}
