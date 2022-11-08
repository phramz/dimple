Dimple (DI)
===========

Lightweight dependency injection container library for Golang inspired by [Pimple](https://github.com/silexphp/Pimple).

## Install

```shell
go get github.com/phramz/dimple
```
## Usage

Here a very basic example how it works in general:
```go
func main() {
	// Builder() returns an instance of ContainerBuilder to configure the container
	builder := dimple.Builder(
		// let's add our favorite time format as parameter to the container so other
		// services can pick it up
		dimple.Param("config.time_format", time.Kitchen),

		// we can just add an anonymous function as factory for our "logger" service since it does not
		// depend on other services
		// and therefore does not need any context
		dimple.Service("logger", dimple.WithFn(func() any {
			logger := logrus.New()
			logger.SetOutput(os.Stdout)

			return logger
		})),

		// this service depends on the "logger" to output the time and "config.time_format" for the
		// desired format.
		// that is why we need to use WithContextFn to get to the container and context
		dimple.Service("service.time", dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
			// you get can whatever dependency you need from the container as
			// long as you do not create a circular dependency
			logger := ctx.Container().MustGet("logger").(*logrus.Logger)
			format := ctx.Container().MustGet("config.time_format").(string)

			return &TimeService{
				logger: logger.WithField("service", ctx.ServiceID()),
				format: format,
			}, nil
		})),
	)

	// once we're done building we can retrieve a new instance of Container
	container, err := builder.Build(context.Background())
	if err != nil {
		panic(err)
	}

	// this is optional, but recommend since it will instantiate all service eager and
	// would return an error if there are issues. We don't want it to panic during runtime but
	// rather error on startup
	if err = container.Boot(); err != nil {
		panic(err)
	}

    // retrieve the *TimeService instance via generic function. Beware that illegal type assertions will panic
    timeService := dimple.MustGetT[*TimeService](container, "service.time")
    timeService.Now()
}
```

Full example see [examples/basic/main.go](./examples/basic/main.go)

### Tags

It is possible to annotate public struct members using the `inject` tag to get all necessary dependencies
without explicitly registering a struct as service and the need of a factory function.

```go

type TaggedTimeService struct {
	Logger *logrus.Logger `inject:"logger"`             // each tag point to the registered definition ID
	Format string         `inject:"config.time_format"` // fields must be public for this to work!
}

func (t *TaggedTimeService) Now() {
	t.Logger.Infof("It is %s", time.Now().Format(t.Format))
}

func main() {
	container := dimple.Builder(
		dimple.Param("config.time_format", time.Kitchen),
		dimple.Service("logger", dimple.WithFn(func() any {
			l := logrus.New()
			l.SetLevel(logrus.DebugLevel)
			return l
		})),
		dimple.Service("service.time", dimple.WithInstance(&TaggedTimeService{})),
	).
		MustBuild(context.Background())

    timeService := dimple.MustGetT[*TaggedTimeService](container, "service.time")
    logger := dimple.MustGetT[*logrus.Logger](container, "logger")

    logger.Debug("timeService says ...")
    timeService.Now()

    // if you want to inject services into struct which is not itself registered as a service
    // you might do it using Inject()
    anotherInstance := &TaggedTimeService{}
    if err := container.Inject(anotherInstance); err != nil {
        panic(err)
    }

    logger.Debug("anotherInstance says ...")
    anotherInstance.Now()
}
```
Full example see [examples/basic/main.go](./examples/basic/main.go)

### Decorators

Decorator can be used to wrap a service with another.

```go

// TimeServiceInterface it might be beneficial to have an interface for abstraction when it comes to service decoration
type TimeServiceInterface interface {
	Now()
}

// OriginTimeService is the service we want to decorate
type OriginTimeService struct {
	Logger *logrus.Logger `inject:"logger"`
	Format string         `inject:"config.time_format"`
}

func (t *OriginTimeService) Now() {
	t.Logger.Infof("It is %s", time.Now().Format(t.Format))
}

// TimeServiceDecorator will decorate OriginTimeService
type TimeServiceDecorator struct {
	Logger    *logrus.Logger `inject:"logger"`
	Decorated TimeServiceInterface
}

func (t *TimeServiceDecorator) Now() {
	t.Logger.Infof("%d seconds elapsed since January 1, 1970 UTC", time.Now().Unix())
	t.Decorated.Now() // let's call the origin TaggedService as well
}

func main() {
	container := dimple.Builder(
		dimple.Param("config.time_format", time.Kitchen),
		dimple.Service("logger", dimple.WithInstance(logrus.New())),
		dimple.Service("service.time", dimple.WithInstance(&OriginTimeService{})),
		dimple.Decorator("service.time.decorator", "service.time", dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
			return &TimeServiceDecorator{
				// we can get the inner (origin) instance if we need to
				Decorated: ctx.Decorated().(TimeServiceInterface),
			}, nil
		})),
	).
		MustBuild(context.Background())

    // now when we call MustGet() we will actually receive the TimeServiceDecorator
	// instead of OriginTimeService
    timeService := dimple.MustGetT[TimeServiceInterface](container, "service.time")
    timeService.Now()
}
```

Full example see [examples/decorator/main.go](./examples/decorator/main.go)

## Build-in services

### Container
You can always get or inject the service container instance by addressing
the service-id `container` e.g.:

```go
type ContainerAwareService struct {
	container Container `inject:"container"`
}
```

### Context
You can always get or inject the context given at Build() by addressing
the service-id `context` e.g.:

```go
type ContextAwareService struct {
	ctx context.Context `inject:"context"`
}
```

## Credits

This library is based on various awesome open source libraries kudos going to:
* https://github.com/silexphp/Pimple
* https://github.com/thoas/go-funk
* https://github.com/stretchr/testify

## License

This project is licensed under MIT License. See [LICENSE](./LICENSE) file.
