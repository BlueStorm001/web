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
	"context"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"log"
	"sync"
	"time"
)

type HttpServer struct {
	Addr      string
	Port      string
	Requester sync.Map
	Handler   func(conn *HttpConn, data []byte)
	Mux       sync.Mutex
	gnet.BuiltinEventEngine
	eng gnet.Engine
}

type HttpConn struct {
	RemoteAddr     string
	ConnectionTime time.Time
	Closed         bool
	ResponseStatus int32
	RequestTime    time.Time
	Field          sync.Map
	RequestNumber  int
	ResponseNumber int
	Path           string
	ContentType    string
	eng            gnet.Conn
}

func (hs *HttpServer) Load(addr string) (*HttpConn, bool) {
	v, ok := hs.Requester.Load(addr)
	if !ok {
		return nil, false
	}
	return v.(*HttpConn), true
}

func (hs *HttpServer) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	log.Println("start http server", hs.Addr, "success")
	return gnet.None
}

func (hs *HttpServer) OnShutdown(eng gnet.Engine) {
	hs.Requester.Range(func(k, v interface{}) bool {
		c := v.(*HttpConn)
		if !c.Closed {
			c.Closed = true
			c.Close()
		}
		return true
	})
}

func (hs *HttpServer) OnOpen(cn gnet.Conn) (out []byte, action gnet.Action) {
	addr := cn.RemoteAddr()
	if addr == nil {
		return
	}
	c := &HttpConn{
		RemoteAddr:     addr.String(),
		ConnectionTime: time.Now(),
		RequestTime:    time.Now(),
		eng:            cn,
	}
	hs.Requester.Store(c.RemoteAddr, c)
	return
}

func exception(s string) {
	if err := recover(); err != nil {
		fmt.Println(s+"异常", err)
	}
}

func (hs *HttpServer) OnClose(cn gnet.Conn, err error) (action gnet.Action) {
	addr := cn.RemoteAddr()
	if addr == nil {
		return
	}
	if c, ok := hs.Load(addr.String()); ok {
		c.Closed = true
	}
	return
}

func (hs *HttpServer) OnTraffic(cn gnet.Conn) gnet.Action {
	addr := cn.RemoteAddr()
	if addr == nil {
		return gnet.Close
	}
	c, ok := hs.Load(addr.String())
	if !ok {
		return gnet.Close
	}
	data, err := cn.Next(-1)
	if err != nil {
		return gnet.Close
	}
	c.RequestTime = time.Now()
	hs.Handler(c, data)
	return gnet.None
}

var httpServer = new(HttpServer)

func start() (err error) {
	defer exception("Serve")
	if err != nil {
		return err
	}
	httpServer.Addr = "tcp://:" + httpServer.Port
	return gnet.Run(httpServer,
		httpServer.Addr,
		gnet.WithLockOSThread(true),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
		gnet.WithReuseAddr(true),
	)
}

func stop() error {
	httpServer.Requester.Range(func(key, value any) bool {
		conn := value.(*HttpConn)
		if conn.Closed {
			return true
		}
		conn.Close()
		return true
	})
	return gnet.Stop(context.Background(), httpServer.Addr)
}

func (c *HttpConn) Write(data []byte) (err error) {
	return c.eng.AsyncWrite(data, nil)
}

func (c *HttpConn) Close() error {
	return c.eng.Close()
}
