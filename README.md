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
    const (
        ServiceLogger      = "logger"
        ServiceTimeService = "service.time"
        ParamTimeFormat    = "config.time_format"
    )

    container := dimple.New(context.Background()).
        // let's add our favorite time format as parameter to the container so other
		// services can pick it up.
        Add(dimple.Param(ParamTimeFormat, time.Kitchen)).

        // we can just add an anonymous function as factory for our "logger" service
		// since it does not depend on other services
        // and therefore does not need any context
        Add(dimple.Service(ServiceLogger, dimple.WithFn(func() any {
            logger := logrus.New()
            logger.SetOutput(os.Stdout)

            return logger
        }))).

        // this service depends on the "logger" to output the time and "config.time_format" for the desired format.
        // that is why we need to use WithContextFn to get to the container and context
        Add(dimple.Service(ServiceTimeService, dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
            logger := ctx.Container().Get(ServiceLogger).(*logrus.Logger)
            format := ctx.Container().Get(ParamTimeFormat).(string)

            return &TimeService{
                logger: logger.WithField("service", ctx.ServiceID()),
                format: format,
            }, nil
        })))

    timeService := container.Get(ServiceTimeService).(*TimeService)
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
	container := dimple.New(context.Background()).
		Add(dimple.Param(ParamTimeFormat, time.Kitchen)).
		Add(dimple.Service(ServiceLogger, dimple.WithInstance(logrus.New()))).
		Add(dimple.Service(ServiceTimeService, dimple.WithInstance(&TaggedTimeService{})))

	go func() {
		timeService := container.Get(ServiceTimeService).(*TaggedTimeService)

		for {
			select {
			case <-container.Ctx().Done():
				return
			default:
				// we will output the time every second
				time.Sleep(time.Second)
				timeService.Now()

				// if you want to inject services into struct
				// which is not itself registered use Inject()
				anotherInstance := &TaggedTimeService{}
				if err := container.Inject(anotherInstance); err != nil {
					panic(err)
				}

				anotherInstance.Now()
			}
		}
	}()

	<-container.Ctx().Done()
}

```
Full example see [examples/basic/main.go](./examples/basic/main.go)

### Decorators

Decorator can be used to wrap a service with another.

```go

func main() {
	container := dimple.New(context.Background()).
		Add(dimple.Param(ParamTimeFormat, time.Kitchen)).
		Add(dimple.Service(ServiceLogger, dimple.WithInstance(logrus.New()))).
		Add(dimple.Service(ServiceTime, dimple.WithInstance(&OriginTimeService{}))).
		Add(dimple.Decorator(ServiceTimeDecorator, ServiceTime, dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
			return &TimeServiceDecorator{
				// we can get the inner (origin) instance if we need to
				Decorated: ctx.Decorated().(TimeServiceInterface),
			}, nil
		})))

	go func() {
		// now when we Get() the ServiceTime, we will actually receive the TimeServiceDecorator
		timeService := container.Get(ServiceTime).(TimeServiceInterface)

		for {
			select {
			case <-container.Ctx().Done():
				return
			default:
				// we will output the time every second
				time.Sleep(time.Second)
				timeService.Now()
			}
		}
	}()

	<-container.Ctx().Done()
}
```

Full example see [examples/decorator/main.go](./examples/decorator/main.go)

## Credits

This library is based on various awesome open source libraries kudos going to:
* https://github.com/silexphp/Pimple
* https://github.com/thoas/go-funk
* https://github.com/stretchr/testify

## License

```
MIT License

Copyright (c) 2022 Maximilian Reichel

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
