package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/rfraposa/weatherbeat/beater"
)

func main() {
	err := beat.Run("weatherbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
