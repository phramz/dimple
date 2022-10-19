package main

import (
	"context"
	"os"
	"time"

	"github.com/phramz/dimple"
	"github.com/sirupsen/logrus"
)

const (
	ServiceLogger      = "logger"
	ServiceTimeService = "service.time"
	ParamTimeFormat    = "config.time_format"
)

// go run main.go
func main() {
	container := dimple.New(context.Background()).
		Add(dimple.Param(ParamTimeFormat, time.Kitchen)).
		Add(dimple.Service(ServiceLogger, dimple.WithFn(func() any {
			logger := logrus.New()
			logger.SetOutput(os.Stdout)

			return logger
		}))).

		// same as in the basic example but we do not have a factory and let do all the injection based on the tagged
		// service id (see below)
		// Any instance in the container will be checked for dimple tags implicitly. If you want to use injection for plain
		// structs that are not defined as service with in the container you can as well do it by e.g.:
		//
		// services := &struct {
		// 		Logger          *logrus.Logger `dimple:"logger"`
		// 		TimeService     *TimeService   `dimple:"service.time"`
		// 		ParamTimeFormat string         `dimple:"config.time_format"`
		// }{}
		//
		// 	if err := container.Inject(services); err != nil {
		//		panic(err)
		//	}
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
			}
		}
	}()

	<-container.Ctx().Done()
}

type TaggedTimeService struct {
	Logger *logrus.Logger `dimple:"logger"`             // each tag point to the registered definition ID
	Format string         `dimple:"config.time_format"` // fields must be public for this to work!
}

func (t *TaggedTimeService) Now() {
	t.Logger.Infof("It is %s", time.Now().Format(t.Format))
}
