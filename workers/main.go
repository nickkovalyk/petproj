package workers

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Invoice InvoiceConfig
}
type Job interface {
	Execute()
}

type Worker struct {
	ID   int
	Jobs chan Job
	die  chan struct{}
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case job, ok := <-w.Jobs:
				if !ok {
					return
				}
				job.Execute()
			case <-w.die:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	logrus.Infof("worker [%d] is stopping", w.ID)
	w.die <- struct{}{}
}
