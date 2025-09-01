uber/fx 框架

[fx](https://link.juejin.cn?target=https%3A%2F%2Fgithub.com%2Fuber-go%2Ffx) 是 uber 2017 年开源的依赖注入解决方案，不仅仅支持常规的依赖注入，还支持生命周期管理。

从官方的视角看，fx 能为开发者提供的三大优势：

- 代码复用：方便开发者构建松耦合，可复用的组件；
- 消除全局状态：Fx 会帮我们维护好单例，无需借用 `init()` 函数或者全局变量来做这件事了；
- 经过多年 Uber 内部验证，足够可信。



#####  核心概念

# provider 声明依赖关系

Provider 就是我们常说的构造器，能够提供对象的生成逻辑。在 Fx 启动时会创建一个容器，我们需要将业务的构造器传进来，作为 Provider。类似下面这样：

```
app = fx.New(
   fx.Provide(NewService),
   fx.Provide(NewBusiness),
   fx.Provide(NewDatabase),
)
```

这里面的 newXXX 函数，就是我们的构造器，类似这样：

```
func NewDatabase() *Database {
	return &Database{}
}
```

# invoker 应用的启动器

provider 是懒加载的，仅仅 Provide 出来我们的构造器，是不会当时就触发调用的，而 invoker 则能够直接触发业务提供的函数运行。

```
fx.Invoke(func(svc service.Service) {
			client := NewClient(svc)
			fmt.Println(client.MakeRequest())
		}),
```



# module 模块化组织依赖

```
func Module(name string, opts ...Option) Option {
	mo := moduleOption{
		name:    name,
		options: opts,
	}
	return mo
}
```

fx 中的 module 也是经典的概念。实际上我们在进行软件开发时，分层分包是不可避免的。而 fx 也是基于模块化编程。使用 module 能够帮助我们更方便的管理依赖：

```
app := fx.New(
		serviceModule,
		businessModule,
		databaseModule,

		fx.Invoke(func(svc service.Service) {
			client := NewClient(svc)
			fmt.Println(client.MakeRequest())
		}),

		//fx.NopLogger, // 如果需要日志，可以启用默认日志或其他日志实现
	)
```



