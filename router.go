package httprouter

import (
	"net/http"
)

// Handle 是一个funcation用来注册到路由并处理HTTP请求
type Handle func(http.ResponseWriter, *http.Request, Params)

// Params 是一个单一的URL参数，包含一个key和value
type Params struct {
	key   string
	value string
}

//路由器是一种http.Handler可用于调度，通过可配置的请求路由到不同的处理函数
type Router struct {
	trees map[string]*node
	// 启用如果当前路线不匹配则自动重定向，但处理程序有（无）结尾的斜线存在的路径。
	// 例如，如果/foo/所请求的路由只能存在于/foo，那么客户端被重定向到/foo的，
	// HTTP状态码301 GET请求和307的所有其他请求方法。
	RedirectTrailingSlash bool

	// 如果启用，路由器会尝试修复当前请求的路径，如果没有处理器注册。像../或//这种多余的
	// 路径元素将被删除。随后路由器将对处理后的路径进行匹配，不区分大小。如果找到一个处理器
	// 对应此路由，路由器使得重定向到与状态码301为GET请求和307的所有其他的请求方法校
	// 正的路径。例如/foo和/..//Foo可能被重定向到/foo中。
	// RedirectTrailingSlash是独立于此选项的。
	RedirectFixedPath bool

	// 如果启用，则路由器将检查是否有另一种方法来处理当前路由，如果当前的请求不能被路由匹配。
	// 这种情况情况下，请求被返回"Method Not Allowed"和HTTP状态代码405，如果没有其他的
	// 方法可以处理路由，该请求被委托给NOTFOUND 处理器来处理。
	HandleMethodNotAllowed bool

	// 配置一个处理器，在路由没有匹配上是被调用，如果没有配置则直接使用http.Handler
	NotFound http.Handler

	// 配置一个http.Handler，当路由没有配置上并且HandleMethodNotAllowed被设置为ture事调用
	// 如果没有设置，将使用http.Error with http.StatusMethodNotAllowed
	MethodNotAllowed http.Handler

	// 函数被用来作为painc的恢复处理程序。它应该来产生一个错误页面，并返回HTTP错误
	// 代码500（内部服务器错误）。该处理器可以被用来保持你的服务器正常，如果有未恢复的panic。
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// New 返回一个新的路由实例
// 路径自动修正，包括尾部斜杠处理，默认被启用
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
	}
}

// Handle 注册请求处理器，用来处理GET,POST,PUT,PATCH和DELETE请求。这个功能可以处理非标准
// 定制方法(例如，内部通讯和代理)
func (r *Router) Handle(method, path string, handle Handle) {
	if path[0] != '/' {
		panic("路径必须以'/'开头 '" + path + "'")
	}

	// 初始化trees
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

}
