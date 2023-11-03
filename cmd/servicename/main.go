package main

// TODO: Change your package import names
// (e.g. from mssfoobar/aoh-service-template/... to /mssfoobar/aoh-solve-all-your-problems/...)
import (
	"flag"
	"mssfoobar/aoh-service-template/pkg/servicename"
	"mssfoobar/aoh-service-template/pkg/utils"
	"os"
	"os/signal"
	"syscall"
)

// Entrypoint for executable Golang programme
func main() {
	var confFile string
	flag.StringVar(&confFile, "c", "./.env", "environment file")
	flag.Parse()

	if confFile == "" {
		flag.PrintDefaults()
		return
	}

	conf := utils.Config{}
	conf.Load(confFile)

	w := servicename.New()
	w.Start(conf)

	defer w.Stop()

	// Press Ctrl+C to exit the process
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)
	<-quitCh
}
