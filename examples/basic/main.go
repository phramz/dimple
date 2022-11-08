package main

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/phramz/dimple"
)

// each definition in the container need an ID so it could be referenced. It might be a good idea to introduce
// constants for this since it could make your life easier when it comes to refactoring stuff.
// though any string is fine ... make sure that its somehow unique otherwise you might overwrite another
// by accident definition with the same id
const (
	ServiceLogger   = "logger"
	ServiceTime     = "service.time"
	ParamTimeFormat = "config.time_format"
)

// go run main.go
func main() {
	// with the Builder we can configure the container
	builder := dimple.Builder(
		// let's add our favorite time format as parameter to the container so other services can pick it up
		dimple.Param(ParamTimeFormat, time.Kitchen),

		// we can just add an anonymous function as factory for our "logger" service since it does not depend on other services
		// and therefore does not need any context
		dimple.Service(ServiceLogger, dimple.WithFn(func() any {
			logger := logrus.New()
			logger.SetOutput(os.Stdout)

			return logger
		})),

		// this service depends on the "logger" to output the time and "config.time_format" for the desired format.
		// that is why we need to use WithContextFn to get to the container and context
		dimple.Service(ServiceTime, dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
			// you get can whatever dependency you need from the container as
			// long as you do not create a circular dependency
			logger := dimple.MustGetT[*logrus.Logger](ctx.Container(), ServiceLogger)
			format := dimple.MustGetT[string](ctx.Container(), ParamTimeFormat)

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

	// this is not necessary, but recommend since it will instantiate all service eager and
	// would return an error if there are issues. We don't want it to panic during runtime but rather error on startup
	if err = container.Boot(); err != nil {
		panic(err)
	}

	go func() {
		// retrieve the *TimeService instance via generic function. Beware that illegal type assertions will panic
		timeService := dimple.MustGetT[*TimeService](container, ServiceTime)

		// if you want to handle error you might consider this approach:
		//
		// val, err := container.Get(ServiceTime)
		// if err != nil {
		// 		panic("cannot instantiate service")
		// }
		// timeService, ok := val.(*TimeService)
		// if !ok {
		// 		panic("illegal type assertion")
		// }

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

type TimeService struct {
	logger *logrus.Entry
	format string
}

func (t *TimeService) Now() {
	t.logger.Infof("It is %s", time.Now().Format(t.format))
}
