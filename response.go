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
	"strconv"
	proto3 "web/proto"
)

type Response struct {
	StatusCode     string
	AcceptEncoding string
	ContentType    string
	Header         []string
	Body           []byte
	GzipBodySize   int64
}

// responseBody
//1xx：指示信息--表示请求已接收，继续处理
//2xx：成功--表示请求已被成功接收、理解、接受
//3xx：重定向--要完成请求必须进行更进一步的操作
//4xx：客户端错误--请求有语法错误或请求无法实现
//5xx：服务器端错误--服务器未能实现合法的请求
func (ws *Server) responseBody(request proto3.Request) []byte {
	if request.HttpStatusCode == "" {
		request.HttpStatusCode = "200"
	}
	var b []byte
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, request.HttpStatusCode...)
	b = append(b, ' ')
	b = append(b, "OK"...)
	b = append(b, '\r', '\n')

	//if request.Timestamp > 0 {
	//	b = append(b, "Net-Time-Consuming: "...)
	//	b = append(b, fmt.Sprint(time.Since(time.UnixMilli(request.Timestamp)))...)
	//	b = append(b, '\r', '\n')
	//}
	//b = append(b, "Date: "...)
	//b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	//b = append(b, '\r', '\n')

	b = append(b, "Content-Type: "...)
	b = append(b, request.ContentType...)
	b = append(b, '\r', '\n')
	//Header
	for _, header := range request.Header {
		b = append(b, header...)
		b = append(b, '\r', '\n')
	}
	bodyLength := int64(len(request.Body))
	if ws.AutomaticCompression && bodyLength >= ws.AutomaticCompressionSize {
		request.Body = GzipCompressBytes(request.Body)
		bodyLength = int64(len(request.Body))
		b = append(b, "Content-Encoding: gzip"...)
		b = append(b, '\r', '\n')
	}
	//length
	b = append(b, "Content-Length: "...)
	b = strconv.AppendInt(b, bodyLength, 10)
	//body
	b = append(b, '\r', '\n')
	b = append(b, '\r', '\n')
	if bodyLength > 0 {
		b = append(b, request.Body...)
	}
	return b
}
