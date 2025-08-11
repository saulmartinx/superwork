package main

import (
	"log"
	"strconv"
	"time"
)

func must(b []byte, err error) []byte {
	if err != nil {
		log.Panic(err)
	}
	return b
}

func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(i/1000, 0)

	return &t, nil
}
