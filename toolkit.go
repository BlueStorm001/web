package web1

import (
	"bytes"
	"compress/gzip"
)

type Query struct {
	RawUrl  string
	Path    string
	File    string
	Suffix  string
	Get     map[string]string
	Anchors []string
	n       int
	step    int
}

func QueryPath(path string) string {
	data := &bytes.Buffer{}
stop:
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '?', '#':
			break stop
		default:
			data.WriteByte(c)
		}
	}
	return data.String()
}

func QueryForm(path string) *Query {
	q := &Query{RawUrl: path}
	data := &bytes.Buffer{}
	l := len(q.RawUrl)
stop:
	for i := 0; i < l; i++ {
		c := q.RawUrl[i]
		switch c {
		case '?', '#':
			q.n = i + 1
			if c == '#' {
				q.step = 1
			}
			break stop
		default:
			data.WriteByte(c)
		}
	}
	q.Path = data.String()
	data.Reset()
	if q.n > 0 {
		q.Get = make(map[string]string)
		var keyStr string
		for i := q.n; i < l; i++ {
			c := path[i]
			switch c {
			case '&', '#':
				switch q.step {
				case 0:
					q.Get[keyStr] = data.String()
				case 1:
					q.Anchors = append(q.Anchors, data.String())
				}
				switch c {
				case '#':
					q.step = 1
				default:
					q.step = 0
				}
				data.Reset()
				keyStr = ""
			case '=':
				if keyStr != "" {
					data.WriteByte(c)
				} else {
					keyStr = data.String()
					data.Reset()
				}
			default:
				data.WriteByte(c)
			}
		}
		switch q.step {
		case 1:
			q.Anchors = append(q.Anchors, data.String())
		default:
			if keyStr != "" {
				q.Get[keyStr] = data.String()
			}
		}
		data.Reset()
	}
	for i := 0; i < len(q.Path); i++ {
		c := q.Path[i]
		switch c {
		case '.':
			q.File = data.String()
			data.Reset()
		case '/':
			data.Reset()
		default:
			data.WriteByte(c)
		}
	}
	if q.File == "" {
		q.File = data.String()
	} else {
		q.Suffix = data.String()
	}
	data.Reset()
	return q
}

// GzipCompressBytes Gzip压缩 bytes
func GzipCompressBytes(input []byte) []byte {
	var buf = &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	leng, err := w.Write(input)
	if err != nil || leng == 0 {
		return nil
	}
	err = w.Flush()
	if err != nil {
		return nil
	}
	err = w.Close()
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
