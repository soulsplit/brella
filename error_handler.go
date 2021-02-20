package main

import (
	"fmt"
	"os"
	"time"
)

func retryOnError(err error) {
	for errorCounter := 0; errorCounter <= 6; errorCounter++ {
		if err != nil {
			fmt.Print(err)
			if errorCounter > 5 {
				os.Exit(1)
			}
			time.Sleep(60 * time.Second)
		}
	}
}