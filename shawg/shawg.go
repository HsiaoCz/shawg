package shawg

import (
	"fmt"
	"net/http"
)

// 定义路由处理函数的模板
type HandlerFunc func(http.ResponseWriter, *http.Request)

// 定义类似于gin的引擎
type Engine struct {
	router map[string]HandlerFunc
}

// 初始化引擎函数
func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

// 添加路由方法
func (e *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	e.router[key] = handler
}

// get方法
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

// POST方法
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

// 启动方法
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

// 这里实现ServeHTTP方法
// 代表我们的引擎可以作为一个handler
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + "-" + r.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, r)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND:%s\n", r.URL)
	}
}
