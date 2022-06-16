package common

import (
	"log"
	"os"
)

var (
	Gslog = log.New(os.Stdout, "[GS Log]", log.Lshortfile)
)
