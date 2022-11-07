package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/phramz/dimple"
)

const (
	ServiceLogger        = "logger"
	ServiceTime          = "service.time"
	ServiceTimeDecorator = "service.time.decorator"
	ParamTimeFormat      = "config.time_format"
)

type TimeServiceInterface interface {
	Now()
}

type OriginTimeService struct {
	Logger *logrus.Logger `inject:"logger"`
	Format string         `inject:"config.time_format"`
}

func (t *OriginTimeService) Now() {
	t.Logger.Infof("It is %s", time.Now().Format(t.Format))
}

type TimeServiceDecorator struct {
	Logger    *logrus.Logger `inject:"logger"`
	Decorated TimeServiceInterface
}

func (t *TimeServiceDecorator) Now() {
	t.Logger.Infof("%d seconds elapsed since January 1, 1970 UTC", time.Now().Unix())
	t.Decorated.Now() // let's call the origin TaggedService as well
}

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
