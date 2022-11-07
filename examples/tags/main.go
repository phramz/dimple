package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/phramz/dimple"
)

const (
	ServiceLogger   = "logger"
	ServiceTime     = "service.time"
	ParamTimeFormat = "config.time_format"
)

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
		Add(dimple.Service(ServiceTime, dimple.WithInstance(&TaggedTimeService{})))

	go func() {
		timeService := container.Get(ServiceTime).(*TaggedTimeService)

		for {
			select {
			case <-container.Ctx().Done():
				return
			default:
				// we will output the time every second
				time.Sleep(time.Second)
				timeService.Now()

				// if you want to inject services into struct which is not itself registered
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
