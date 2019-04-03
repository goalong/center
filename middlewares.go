package center

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)



// 将多个中间件连接起来
func Chain(f Handler, middlewares ...Middleware) Handler {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

// 日志中间件，打印URL及执行时间
func Logging() Middleware {
	return func(f Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request, params url.Values) {
			fmt.Println("logging")
			start := time.Now()
			defer func() { log.Println(r.RequestURI, time.Since(start)) }()
			f(w, r, params)
		}
	}
}

func Recover() Middleware {
	return func(f Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request, params url.Values) {
			fmt.Println("doing recover")
			defer func() {
				if err := recover(); err != nil {
					log.Println("catch panic")
					log.Println("panic: ", err)
				}
			}()
			f(w, r, params)

		}
	}
}

// 设置返回json类型
func ReturnJson() Middleware {
	return func(f Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request, params url.Values) {
			w.Header().Set("content-type", "application/json")
			f(w, r, params)
		}
	}
}
