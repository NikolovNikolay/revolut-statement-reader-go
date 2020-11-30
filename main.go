package main

import (
	"fmt"
	"github.com/nikolovnikolay/revolut-statement-reader-go/internal/actions"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "folder",
				Aliases:  []string{"f"},
				Usage:    "Load report statements from `FOLDER`",
				Required: true,
			},
		},
		Action: actions.ProcessStatements,
	}

	setLogger()
	setAppInfo(app)
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Process could not complete successfully")
		log.Fatal(err)
	}
}

func setAppInfo(app *cli.App) {
	app.Name = "Tax calculator"
	app.Usage = "Tax calculator for Revolut and eToro activities"
	app.Authors = []*cli.Author{
		{
			Name:  "Nikolay Nikolov",
			Email: "nikolov89@gmail.com",
		},
	}
	app.Version = "0.0.1"
}

func setLogger() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
