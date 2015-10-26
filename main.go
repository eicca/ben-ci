package main

import (
	"os"

	"github.com/eicca/ben-ci/ci"
	"github.com/eicca/ben-ci/web"
)

func main() {
	go web.ListenAndServe()
	ci.ListenAndServe(":" + os.Getenv("LISTENER_PORT"))
}
