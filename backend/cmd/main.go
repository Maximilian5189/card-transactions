package main

import (
	"backend/backup"
	"backend/logger"
	"flag"
	"fmt"
)

func main() {
	offset := flag.Int("offset", 0, "offset")

	flag.Usage = func() {
		fmt.Println("Usage: cmd [options]")
		flag.PrintDefaults()
	}

	flag.Parse()

	logger := logger.NewLogger()

	b, err := backup.New(logger)

	if err != nil {
		logger.Error(fmt.Sprintf("error instantiating backup: %s ", err))
	} else {
		b.Download("/app/database/database.db", *offset)
	}
}
