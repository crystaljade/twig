# twig
twig 是一个面向webapi的简单的webserver （目前预览发布，API没有冻结，请注意）

> Twig 采用QQ群提供技术支持，QQ群号：472900117

## 入门

```go
package main

import (
	"github.com/twiglab/twig"
	"github.com/twiglab/twig/middleware"
)

func main() {
	api := twig.Default()

	twig.Config(api).
		Use(middleware.Recover()).
		Get("/hello", twig.HelloTwig).
		Done()

	api.Start()

	twig.Signal(twig.Quit())
}
```

- Twig的默认监听端口是4321, 可以通过twig.DefaultAddress全局变量修改(`位于var.go中`)，或者自定义自己的Server
- 使用twig.Default()创建`默认的`Twig，默认的Twig包括默认的HttpServer（DefaultServant）,默认的路由实现（RadixTree)，默认的Logger和默认的HttpErrorHandler
- twig.Config是Twig提供的配置工具（注意：在Twig的世界里，工具的定义为外围附属品，并不是Twig必须的一部分），Twig没有像别的webserver一样提供GET，POST等方法，所有的配置工作都通过Config完成
- Twig要求所有的Server的实现必须是*非堵塞*的，Start方法将启动Twig，Twig提供了Signal组件用于堵塞应用，处理系统信号，完成和shell的交互

### Twig没有提供的功能

- Twig没有内置提供Render，或者模板输出功能，Twig专注api
- Twig没有提供Binder功能，Twig认为绑定参数到对象是应用需要考虑的问题
- Twig默认的Server不支持SSL，如何设计并提供好的Server是应用开发者考虑的问题（`实现Server接口，位于serve.go中`） 
- Twig没有提供GET,POST,DELETE等路由方法，但Twig提供了Register和Mounter接口，方便用户分离路由配置工作，并实现模块化配置

Twig最大的特点是简洁，至此讲述的内容，已经足够让您运行并使用Twig。*祝您使用Twig愉快！*

----

## Twig的结构
