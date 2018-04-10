package main

import (
	"log"

	"github.com/dacoursey/whatsmyname/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
