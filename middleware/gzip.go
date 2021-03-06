package middleware

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

type (
	// gzipResponse 定义Gzip响应，实现ResponseWriter接口
	gzipResponse struct {
		eudore.ResponseWriter
		writer *gzip.Writer
	}
	// Gzip 定义gzip压缩处理者。
	Gzip struct {
		pool sync.Pool
	}
)

// NewGzip 创建一个gzip压缩处理者。
func NewGzip(level int) *Gzip {
	h := &Gzip{
		pool: sync.Pool{
			New: func() interface{} {
				gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
				if err != nil {
					return err
				}
				return &gzipResponse{
					writer: gz,
				}
			},
		},
	}
	return h
}

// NewGzipFunc 函数返回一个gzip处理函数。
func NewGzipFunc(level int) eudore.HandlerFunc {
	return NewGzip(level).HandleHTTP
}

// HandleHTTP 方法定义eudore请求处理函数。
func (g *Gzip) HandleHTTP(ctx eudore.Context) {
	// 检查是否使用Gzip
	if !shouldCompress(ctx) {
		ctx.Next()
		return
	}
	// 初始化ResponseWriter
	w, err := g.NewGzipResponse(ctx.Response())
	if err != nil {
		// 初始化失败，正常写入
		ctx.Error(err)
		ctx.Next()
		return
	}
	ctx.SetResponse(w)
	// 设置gzip header
	ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
	ctx.SetHeader(eudore.HeaderVary, eudore.HeaderAcceptEncoding)
	// Next
	ctx.Next()
	w.writer.Close()
	// 回收GzipResponse
	g.pool.Put(w)

}

// NewGzipResponse 创建一个gzip响应。
func (g *Gzip) NewGzipResponse(w eudore.ResponseWriter) (*gzipResponse, error) {
	switch val := g.pool.Get().(type) {
	case *gzipResponse:
		val.ResponseWriter = w
		val.writer.Reset(w)
		return val, nil
	case error:
		return nil, val
	}
	return nil, fmt.Errorf("Create gzipResponse exception")
}

// Write 实现ResponseWriter中的Write方法。
func (w *gzipResponse) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

// Flush 实现ResponseWriter中的Flush方法。
func (w *gzipResponse) Flush() {
	w.writer.Flush()
	w.ResponseWriter.Flush()
}

func shouldCompress(ctx eudore.Context) bool {
	h := ctx.Request().Header
	if !strings.Contains(h.Get(eudore.HeaderAcceptEncoding), "gzip") ||
		strings.Contains(h.Get(eudore.HeaderConnection), "Upgrade") ||
		strings.Contains(h.Get(eudore.HeaderContentType), "text/event-stream") {

		return false
	}

	h = ctx.Response().Header()
	if strings.Contains(h.Get(eudore.HeaderContentEncoding), "gzip") {
		return false
	}

	extension := filepath.Ext(ctx.Path())
	if len(extension) < 4 { // fast path
		return true
	}

	switch extension {
	case ".png", ".gif", ".jpeg", ".jpg":
		return false
	default:
		return true
	}
}

// Push initiates an HTTP/2 server push.
// Push returns ErrNotSupported if the client has disabled push or if push
// is not supported on the underlying connection.
func (w *gzipResponse) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.Push(target, setAcceptEncodingForPushOptions(opts))
}

// setAcceptEncodingForPushOptions sets "Accept-Encoding" : "gzip" for PushOptions without overriding existing headers.
func setAcceptEncodingForPushOptions(opts *http.PushOptions) *http.PushOptions {

	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				eudore.HeaderAcceptEncoding: []string{"gzip"},
			},
		}
		return opts
	}

	if opts.Header == nil {
		opts.Header = http.Header{
			eudore.HeaderAcceptEncoding: []string{"gzip"},
		}
		return opts
	}

	if encoding := opts.Header.Get(eudore.HeaderAcceptEncoding); encoding == "" {
		opts.Header.Add(eudore.HeaderAcceptEncoding, "gzip")
		return opts
	}

	return opts
}
