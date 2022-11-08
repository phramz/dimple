package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/phramz/dimple"
)

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

	go func() {
		// now when we MustGet() the ServiceTime, we will actually receive the TimeServiceDecorator
		timeService := dimple.MustGetT[TimeServiceInterface](container, "service.time")

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
