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

func (ws *Server) httpRequest(data []byte) (request proto3.Request) {
	var (
		line int8
		step int8
		idx  int
		body []byte
	)
span:
	for i := 0; i < len(data); i++ {
		b := data[i]
		switch b {
		case ' ':
			if line == 0 {
				step++
				switch step {
				case 1:
					request.Method = string(body)
				case 2:
					request.Path = string(body)
				}
				body = body[:0]
				continue
			}
		case '\r':
			continue
		case '\n':
			if len(body) == 0 {
				idx = i + 1
				break span
			}
			if line == 0 {
				request.Proto = string(body)
			} else {
				var (
					key   string
					value []byte
				)
				for _, h := range body {
					switch h {
					case ':':
						key = string(value)
						value = value[:0]
						continue
					case ' ':
						if len(value) == 0 {
							continue
						}
					}
					value = append(value, h)
				}
				switch key {
				case "Content-Length", "content-length":
					n, _ := strconv.Atoi(string(value))
					request.ContentLength = int32(n)
				case "Host", "host":
					request.Host = string(value)
				case "Content-Type", "content-type":
					request.ContentType = string(value)
				default:
					if !ws.NoHeader {
						request.Header = append(request.Header, key+":"+string(value))
					}
				}
			}
			body = body[:0]
			line++
			continue
		}
		body = append(body, b)
	}
	if idx > 0 {
		request.Body = data[idx:]
	}
	return request
}
