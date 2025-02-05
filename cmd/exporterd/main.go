package main

import (
	"log/slog"
	"os"

	"github.com/openinnovationai/k8s-amd-exporter/internal/application"
)

func main() {
	app := application.NewApplication()

	if err := app.Run(); err != nil {
		slog.Error("unable to run service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("the application has been finalized")
}
