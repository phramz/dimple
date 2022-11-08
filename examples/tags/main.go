package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/phramz/dimple"
)

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

	go func() {
		timeService := dimple.MustGetT[*TaggedTimeService](container, "service.time")
		logger := dimple.MustGetT[*logrus.Logger](container, "logger")

		for {
			select {
			case <-container.Ctx().Done():
				return
			default:
				// we will output the time every second
				time.Sleep(time.Second)
				logger.Debug("timeService says ...")
				timeService.Now()

				// if you want to inject services into struct which is not itself registered
				anotherInstance := &TaggedTimeService{}
				if err := container.Inject(anotherInstance); err != nil {
					panic(err)
				}

				logger.Debug("anotherInstance says ...")
				anotherInstance.Now()
			}
		}
	}()

	<-container.Ctx().Done()
}
