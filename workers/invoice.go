package workers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/models"
	"gitlab.com/i4s-edu/petstore-kovalyk/services/storage"
	"gitlab.com/i4s-edu/petstore-kovalyk/utils"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/i4s-edu/petstore-kovalyk/db/mappers"
)

const invoiceTemplatePath = "resources/templates/invoice.tmpl"
const tempDirectory = "tmp"

type InvoiceConfig struct {
	Count    int
	Interval utils.Duration
}

type InvoiceJob struct {
	DB       *sqlx.DB
	Interval time.Duration
}

func (i InvoiceJob) Execute() {
	logrus.Info("invoice job start")
	now := time.Now()
	point := now.Unix() - int64(i.Interval.Seconds())
	orders, err := mappers.OrderMapper{DB: i.DB}.GetOldest(point)
	if err != nil {
		logrus.Error(err)
		return
	}
	if len(orders) == 0 {
		logrus.Error("SKIPPING JOB: No orders for that period")
		return
	}
	t, err := template.New("invoice.tmpl").ParseFiles(invoiceTemplatePath)
	if err != nil {
		logrus.Error("invoice template parse error: ", err)
		return
	}

	date := now.Format(time.RFC3339)
	var totalQuantity int
	for _, v := range orders {
		totalQuantity += v.Quantity
	}
	data := struct {
		Orders        []*models.Order
		Date          string
		TotalQuantity int
	}{
		Orders:        orders,
		Date:          date,
		TotalQuantity: totalQuantity,
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		logrus.Error("invoice template execute error: ", err)
		return
	}
	//
	//newInvoice := &models.Invoice{}
	//newInvoice.Body = tpl.String()
	//newInvoice.CreatedDate = now.Unix()
	//
	//err = mappers.InvoiceMapper{DB: i.DB}.Create(newInvoice)
	//if err != nil {
	//	logrus.Error("create invoice, error: ", err)
	//	return
	//}

	invoiceFilename := fmt.Sprintf("invoice_%v.md", date)
	file, err := ioutil.TempFile(tempDirectory, invoiceFilename)
	if err != nil {
		log.Fatal(err)
	}
	logrus.Info("template output", tpl.String())
	if _, err = fmt.Fprint(file, tpl.String()); err != nil {
		logrus.Errorf("cannot write to temp file: %v, %v", file.Name(), err)
	}

	store := storage.GetStorage()
	err = store.Save("invoices", invoiceFilename, "text/markdown; charset=UTF-8", file.Name())
	if err != nil {
		logrus.Error("invoice save error", err)
	}
	defer func() {
		err := os.Remove(file.Name())
		if err != nil {
			logrus.Error("unable to remove temp invoice file:", file.Name(), err)
		}
	}()
	logrus.Info("Invoice worker finished, created file :", invoiceFilename)

}

type InvoiceJobCollector struct {
	Jobs     chan Job
	Interval time.Duration
	DB       *sqlx.DB
	die      chan struct{}
}

func (i *InvoiceJobCollector) Collect() {
	logrus.Info("invoice collect starts, interval=", i.Interval)
	i.Jobs <- InvoiceJob{Interval: i.Interval, DB: i.DB}
	time.Sleep(i.Interval)
}

func (i *InvoiceJobCollector) Start() {
	logrus.Info("invoice worker started")
	go func() {
		for {
			select {
			case <-i.die:
				return
			default:
				i.Collect()
			}
		}
	}()
}

func (i *InvoiceJobCollector) End() {
	i.die <- struct{}{}
}

func DispatchInvoiceWorker(interval time.Duration, db *sqlx.DB) {
	logrus.Info("dispatch invoice worker")
	jobs := make(chan Job)
	worker := Worker{Jobs: jobs}
	collector := InvoiceJobCollector{Interval: interval, Jobs: jobs, DB: db}
	worker.Start()
	collector.Start()
}
