package main

import "nola-go/internal/app"

func main() {

	nola, err := app.NewNola()
	if err != nil {
		panic(err)
	}

	runErr := nola.Run()
	if runErr != nil {
		panic(runErr)
	}
}
