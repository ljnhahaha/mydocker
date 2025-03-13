package main

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2" // imports as package "cli"
)

const (
	VERSION = "v0.0.0"
	NAME    = "MyDocker"
	USAGE   = "MyDocker is a simple container runtime implementation."
)

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Name = NAME
	app.Usage = USAGE

	app.Commands = []*cli.Command{
		&runCommand,
		&initCommand,
		&commitCommand,
		&listCommand,
		&logCommand,
		&execCommand,
		&stopCommand,
		&removeCommand,
		&networkCommand,
		&startCommand,
	}

	app.Before = func(c *cli.Context) error {
		_ = os.MkdirAll("logs", os.ModePerm)
		file, _ := os.OpenFile("logs/runtime.out", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		log.SetOutput(io.MultiWriter(file, os.Stdout))
		log.SetReportCaller(true)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
