package twig

import (
	"context"
	"net/http"
	"os"
	"sync"
)

type M map[string]interface{}

// Identifier 标识符接口
type Identifier interface {
	ID() string
}

// Attacher 用于设置Twig和组件之间的联系
type Attacher interface {
	Attach(*Twig)
}

// Cycler 设置周期管理
type Cycler interface {
	Start() error
	Shutdown(context.Context) error
}

// Namer 命名接口
type Namer interface {
	SetName(string)
}

// Twig
type Twig struct {
	HttpErrorHandler HttpErrorHandler

	Logger Logger // Logger 组件负责日志输出
	Muxer  Muxer  // Muxer 组件负责路由处理
	Server Server // Server 负责Http请求处理

	Debug bool

	pre []MiddlewareFunc
	mid []MiddlewareFunc

	pool sync.Pool

	plugins map[string]Plugin

	name string
}

// 创建空的Twig
func TODO() *Twig {
	return Default()
}

// 创建默认的Twig
func Default() *Twig {
	t := &Twig{
		Debug: false,
		name:  "main",
	}
	t.pool.New = func() interface{} {
		return t.NewCtx(nil, nil)
	}

	t.
		WithServer(DefaultServant()).
		WithHttpErrorHandler(DefaultHttpErrorHandler).
		WithLogger(newLog(os.Stdout, "twig-log-")).
		WithMuxer(NewRadixTree())

	return t
}

func (t *Twig) WithLogger(l Logger) *Twig {
	t.Logger = l
	attach(l, t)
	return t
}

func (t *Twig) WithHttpErrorHandler(eh HttpErrorHandler) *Twig {
	t.HttpErrorHandler = eh
	return t
}

func (t *Twig) WithMuxer(m Muxer) *Twig {
	t.Muxer = m
	attach(m, t)
	return t
}

func (t *Twig) WithServer(s Server) *Twig {
	t.Server = s
	s.Attach(t)
	return t
}

// Pre 中间件支持， 注意Pre中间件工作在路由之前
func (t *Twig) Pre(m ...MiddlewareFunc) {
	t.pre = append(t.pre, m...)
}

// Twig级中间件支持
func (t *Twig) Use(m ...MiddlewareFunc) {
	t.mid = append(t.mid, m...)
}

// Plugin支持
func (t *Twig) AddPlugin(ps ...Plugin) {
	for _, p := range ps {
		attach(p, t)
		t.plugins[p.ID()] = p
	}
}

//获取Plugin
func (t *Twig) Plugin(id string) Plugin {
	return t.plugins[id]
}

// 实现http.Handler
func (t *Twig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := t.pool.Get().(*ctx) // pool 中获取Ctx
	c.Reset(w, r)            // 重置Ctx，放入当前的Resp和Req , Ctx是可以重用的

	h := Enhance(func(ctx Ctx) error { //注意这里是个闭包，闭包中处理Twig级中间件，结束后处理Pre中间件
		t.Muxer.Lookup(r.Method, GetReqPath(r), r, c) // 路由对当前Ctx实现装配
		handler := Enhance(c.Handler(), t.mid)        // 处理Twig级中间件
		return handler(ctx)
	}, t.pre)

	if err := h(c); err != nil { // 链式调用，如果出错，交给Twig的HttpErrorHandler处理
		t.HttpErrorHandler(err, c)
	}

	t.pool.Put(c) // 交还Ctx，后续复用，Http处理过程结束
}

func (t *Twig) Start() error {
	t.Logger.Println(banner)

	for _, p := range t.plugins {
		if cycler, ok := p.(Cycler); ok {
			cycler.Start()
		}
	}

	return t.Server.Start()
}

func (t *Twig) Shutdown(ctx context.Context) error {
	for _, p := range t.plugins {
		if cycler, ok := p.(Cycler); ok {
			cycler.Shutdown(ctx)
		}
	}

	return t.Server.Shutdown(ctx)
}

// 面向第三方路由，提供Ctx的创建功能
// 注意：Twig 不管理第三方路由使用的Ctx，只负责创建，不负责回收
func (t *Twig) NewCtx(w http.ResponseWriter, r *http.Request) Ctx {
	return &ctx{
		req:     r,
		resp:    NewResponseWarp(w),
		t:       t,
		store:   make(M),
		pvalues: make([]string, MaxParam),
		handler: NotFoundHandler,
	}
}

func (t *Twig) AcquireCtx() Ctx {
	c := t.pool.Get().(*ctx)
	return c
}

func (t *Twig) ReleaseCtx(c Ctx) {
	t.pool.Put(c)
}

func (t *Twig) AddHandler(method, path string, handler HandlerFunc, m ...MiddlewareFunc) Route {
	return t.Muxer.AddHandler(method, path, handler, m...)
}

func (t *Twig) SetName(name string) {
	t.name = name
}

func (t *Twig) Name() string {
	return t.name
}

func (t *Twig) ID() string {
	return "Twig@" + t.name
}

func (t *Twig) Config() *Conf {
	return Config(t).WithNamer(t)
}
