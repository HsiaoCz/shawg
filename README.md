## 简单实现一个 web 框架

首先看一个 go net/http 实现的例子

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

// handler echoes r.URL.Path
func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

// handler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}
```

http.ListenAndServe("listenAddr",nil)
这里的这个 nil，指的是默认的多路复用器
http.DefaultServeMux,它可以接收请求，并对请求进行匹配，本质上是维护了一个 map

http.HandleFunc()接收两个参数，一个是 string，代表路由，一个是 http.Handler，代表了路由对应的处理方法
这个 handler 是一个接口

```go
package http

type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}

func ListenAndServe(address string, h Handler) error
```

凡是实现了 ServeHTTP 方法的实例，都是一个 handler

但是这里有一个疑问，我们自己写的方法
`func helloHandler(w http.ResponseWriter,r *http.Request)`
并没有实现 ServeHTTP(wr)方法，为什么它也是 handler 呢？

原因在于，http.HandleFunc()里面调用了 DefaultServeMux.HandleFunc()方法

```go
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}
```

这个 HandleFunc()方法内部实现:

```go
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(pattern, HandlerFunc(handler))
}
```

它在内部又调用了一个 mux.Handle()

这里面的 HandlerFunc()是一个函数类型
它实现了 serveHTTP()方法，它可以将具有适当签名的函数适配成一个 handler
这里其实是做了一个转换

### 1、实现我们自己的多路复用器

```go
// 这个HandlerFunc用来定义路由映射的处理方法
type HandlerFunc func(http.ResponseWriter, *http.Request)

// 引擎 维护一个router的map
// 这个map的key由请求方法和静态路由地址构成
// 例如GET-/ POST-/ DELETE-/
// 针对相同的路由 请求方法不同也可以映射到不同的处理方法
type Engine struct {
	router map[string]HandlerFunc
}

// 用户调用New得到一个引擎
func New() *Engine {
	return &Engine{
		router: make(map[string]HandlerFunc),
	}
}

// 路由注册的方法
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	engine.router[key] = handler
}

// 当用户调用不同的(*Engine)方法，可以将路由和处理方法的映射注册到Engine的map里
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// 这里简单包装一下http.ListenAndServe()方法
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 这里实现ServeHTTP方法，将注册的路由映射取出，执行处理的函数
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + "-" + r.URL.Path
	if handler, ok := engine.router[key]; ok {
		handler(w, r)
        return
	}
	fmt.Fprintf(w, "404 NOT FOUND:%s\n", r.URL)
}


```

### 2、上下文

这一步将handler封装成context
context中封装包含HTML/json/string函数
```go
// 传递给json的encode()函数的数据
// 本质是一个map
type H map[string]interface{}

// 对context进行疯转
type Context struct {
	// 响应
	Writer     http.ResponseWriter
	// 请求体
	req        *http.Request
	// 路径
	Path       string
	// 请求方法
	Method     string
	// 状态
	StatusCode int
}
// 实例化一个context
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
	}
}
// 获取表单信息的方法
func (c *Context) PostForm(key string) string {
	return c.req.FormValue(key)
}
// 获取查询参数得到方法
func (c *Context) Query(key string) string {
	return c.req.URL.Query().Get(key)
}
// 在响应头写入状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}
// 设置响应的消息格式
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}
// 返回string类型的响应消息
func (c *Context) String(code int, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprint(values...)))
}
// 返回Json类型的响应格式
func (c *Context) JSON(code int, obj ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}
// 返回字节切片类的格式
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}
// 返回HTML格式
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
```

整个context的核心就是一个结构体
不过这个context封装的比较简单
这个context具有一些方法，就是我们比较常用的返回响应的方法

然后将原来的代码进行替换
```go
type HandlerFunc func(*Context)

type Engine struct {
	router *router
}

func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	engine.router.handle(c)
}
```

### 3、前缀数路由

所谓的动态路由，就是一条路由规则匹配某一类型而非某一条固定的路由，例如/hello/:name
可以匹配/hello/xiao /hello/fan

实现动态路由最常用的数据结构，被称为前缀树
每一个节点的所有子节点都拥有相同的前缀

这里让每一段作为前缀树的一个节点，通过树结构查询，如果中间某一层的节点都不满足条件，那么就说明没有匹配到的路由，查询结束

我们实现的路由将具备两个功能
参数匹配：例如/p/:lang/doc，可以匹配/p/c/doc 和/p/go/doc
通配匹配。例如/static/*filepath，例如匹配/static/fav.ioc，也可以匹配/static/js/JQery.js，这种模式用于静态服务器，能够递归匹配子路径
