package shawg

import (
	"log"
	"net/http"
)

// 定义路由处理函数的模板
// type HandlerFunc func(http.ResponseWriter, *http.Request)
type HandlerFunc func(*Context)

// 定义类似于gin的引擎
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

// 分组路由
type RouterGroup struct {
	prefix      string        //存路由的前缀
	middlewares []HandlerFunc // 保存注册在当前分组上的中间件
	parent      *RouterGroup  // 分组的父组件，实现分组的嵌套
	engine      *Engine       // 间接的访问Engine的资源
}

// 初始化引擎函数
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// 添加路由方法
func (e *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
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
	c := newContext(w, r)
	e.router.handle(c)
}
