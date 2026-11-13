package main

import (
	"fmt"
	"os"

	"pos/internal/di"
	"pos/internal/infrastructure/cli"
)

func main() {
	container, err := di.BuildContainer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "собрать DI-контейнер: %v\n", err)
		os.Exit(1)
	}
	if err := container.Invoke(func(app *cli.ConsoleApp) error {
		return app.Run()
	}); err != nil {
		fmt.Fprintf(os.Stderr, "работа приложения: %v\n", err)
		os.Exit(1)
	}
}
