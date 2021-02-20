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

func checkPrompt(err error) {
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
}
