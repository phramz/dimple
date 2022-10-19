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
package example

import (
    "context"
    "github.com/phramz/dimple"
    "github.com/sirupsen/logrus"
    "os"
    "time"
)

func basic() {
    const (
        ServiceLogger      = "logger"
        ServiceTimeService = "service.time"
        ParamTimeFormat    = "config.time_format"
    )

    container := dimple.New(context.Background()).
        // let's add our favorite time format as parameter to the container so other services can pick it up
        Add(dimple.Param(ParamTimeFormat, time.Kitchen)).

        // we can just add an anonymous function as factory for our "logger" service since it does not depend on other services
        // and therefore does not need any context
        Add(dimple.Service(ServiceLogger, dimple.WithFn(func() any {
            logger := logrus.New()
            logger.SetOutput(os.Stdout)

            return logger
        }))).

        // this service depends on the "logger" to out put the time
        // that is why we need to use FactoryFn to get to the container
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
Have a look at the [examples](./examples) folder for full code examples.

### Tags

### Services

### Parameters

### Decorators

## Credits

This library is based on various awesome open source libraries shout-outs to:
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
