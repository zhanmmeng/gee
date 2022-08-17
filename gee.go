package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc  定义了 gee 使用的请求处理程序
// HandlerFunc defines the request handler used by gee
type HandlerFunc func(ctx *Context)

// Engine implement the interface of ServeHTTP
// Engine 实现ServeHTTP的接口
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // store all groups 存储所有组

	htmlTemplates *template.Template //for html render 用于html渲染
	funcMap template.FuncMap //for html render
}

// RouterGroup 路由组
type RouterGroup struct{
	prefix string
	middlewares []HandlerFunc //support middlewares
	parent *RouterGroup // support nesting 支持嵌套
	engine *Engine // all groups share a Engine instance 所有组共享一个Engine实例
}

//New is the constructor of gee.Engine (Engine的构造函数)
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}

	return &Engine{
		router: newRouter(),
	}
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance 记住所有组共享同一个引擎实例
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups,newGroup)
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

// create static handler  创建静态处理程序
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix,relativePath)
	fileServer := http.StripPrefix(absolutePath,http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		// 检查文件是否存在和/或我们是否有权访问它
		if _,err := fs.Open(file);err != nil{
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer,c.Req)
	}
}

// Static serve static files 提供静态文件
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers 注册 GET 处理程序
	group.GET(urlPattern, handler)
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method,pattern,handler)
}

//GET 定义添加get请求的方法(GET defines the method to add GET request)
func (engine *Engine) GET(pattern string,handler HandlerFunc)  {
	engine.addRoute("GET",pattern,handler)
}

//POST 定义添加post请求的方法(POST defines the method to add POST request)
func (engine *Engine) POST(pattern string,handler HandlerFunc)  {
	engine.addRoute("POST",pattern,handler)
}

//Run 定义了启动 http 服务器的方法(Run defines the method to start a http server)
func (engine *Engine) Run(add string) (err error) {
	return http.ListenAndServe(add,engine)
}

func (engine *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var middlewares []HandlerFunc
	for _,group := range engine.groups{
		if strings.HasPrefix(request.URL.Path,group.prefix) {
			middlewares = append(middlewares,group.middlewares...)
		}
	}

	c:=newContext(writer,request)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

// Use is defined to add middleware to the group 将定义的中间件添加到组中
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares,middlewares...)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string)  {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
