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

		// this service depends on the "logger" to output the time and "config.time_format" for the desired format.
		// that is why we need to use WithContextFn to get to the container and context
		Add(dimple.Service(ServiceTime, dimple.WithContextFn(func(ctx dimple.FactoryCtx) (any, error) {
			// you get can whatever dependency you need from the container as
			// long as you do not create a circular dependency
			logger := ctx.Container().Get(ServiceLogger).(*logrus.Logger)
			format := ctx.Container().Get(ParamTimeFormat).(string)

			return &TimeService{
				logger: logger.WithField("service", ctx.ServiceID()),
				format: format,
			}, nil
		})))

	// this is not necessary, but recommend since it will instantiate all service eager and
	// would return an error if there are issues. We don't want it to panic during runtime but rather error on startup
	if err := container.Boot(); err != nil {
		panic(err)
	}

	go func() {
		timeService := container.Get(ServiceTime).(*TimeService)
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
