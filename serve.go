/*
 * Copyright (c) 2021 BlueStorm
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFINGEMENT IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package web1

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	proto3 "web/proto"
)

type Server struct {
	Request                  func(request proto3.Request) *proto3.Request
	Response                 func(request proto3.Request)
	SyncResponse             bool  //同步相应
	AutomaticCompressionSize int64 //压缩大小
	AutomaticCompression     bool  //压缩
	NoHeader                 bool  //不需要请求头
	Timeout                  time.Duration
	TimeoutBody              func(path string) []byte
}

func Service() *Server {
	return &Server{AutomaticCompressionSize: 1024 * 3}
}

// StartRun web server
func (ws *Server) StartRun(port string) {
	if ws.Request == nil {
		fmt.Println("handler must be assigned at startup")
		return
	}
	httpServer.Port = port
	httpServer.Handler = ws.httpHandler
	ws.Response = func(request proto3.Request) {
		if v, ok := httpServer.Requester.Load(request.RemoteAddr); ok {
			ws.responseHandler(request, v.(*HttpConn), false)
		}
	}
	if ws.Timeout > 0 {
		go ws.checkWebRequester()
	}
	go func() {
		err := start()
		if err != nil {
			log.Println("start web server error", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	ws.Stop()
	return
}

func (ws *Server) Stop() {
	msg := "success"
	if err := stop(); err != nil {
		msg = err.Error()
	}
	log.Println("stop http server ", httpServer.Port, msg)
}

var requestAppendKey = "requestAppend"

func (ws *Server) httpHandler(conn *HttpConn, data []byte) {
	var request proto3.Request
	var yes bool
	if v, ok := conn.Field.Load(requestAppendKey); ok {
		yes = true
		request = v.(proto3.Request)
		request.Body = append(request.Body, data...)
	} else {
		request = ws.httpRequest(data)
		request.RemoteAddr = conn.RemoteAddr
	}
	if len(request.Body) < int(request.ContentLength) {
		conn.Field.Store(requestAppendKey, request)
		return
	}
	if yes {
		conn.Field.Delete(requestAppendKey)
	}
	conn.ResponseStatus = 1
	conn.Path = request.Path
	conn.ContentType = request.ContentType
	if ws.SyncResponse {
		ws.responseHandler(request, conn, true)
	} else {
		go ws.responseHandler(request, conn, true)
	}
}

func (ws *Server) responseHandler(request proto3.Request, conn *HttpConn, query bool) {
	if query {
		r := ws.Request(request)
		if r == nil {
			return
		}
		request = *r
	}
	if conn.ResponseStatus == 0 {
		return
	}
	conn.ResponseStatus = 0
	conn.Write(ws.responseBody(request))
}

func (ws *Server) checkWebRequester() {
	for {
		time.Sleep(time.Second * 5)
		now := time.Now() // w.report()
		httpServer.Requester.Range(func(key, value interface{}) bool {
			conn := value.(*HttpConn)
			//清除已关闭连接
			if conn.Closed {
				httpServer.Requester.Delete(key)
			} else {
				if conn.ResponseStatus == 1 && now.Sub(conn.RequestTime) >= ws.Timeout { //超过应答时间
					conn.ResponseStatus = 0
					response := proto3.Request{ContentType: conn.ContentType}
					if ws.TimeoutBody != nil {
						response.Body = ws.TimeoutBody(conn.Path)
						response.HttpStatusCode = "200"
					} else {
						response.HttpStatusCode = "504"
					}
					conn.Write(ws.responseBody(response))
				}
			}
			return true
		})
	}
}
