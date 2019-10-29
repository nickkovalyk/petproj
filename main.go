package main

import (
	_ "github.com/lib/pq"
	"gitlab.com/i4s-edu/petstore-kovalyk/app"
)

const logFile = "logrus.log"

func main() {
	/*file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Problems with %v file %v", logFile, err)
	}
	logrus.SetOutput(file)*/

	serverApp := app.NewApp()
	serverApp.Run()
}
