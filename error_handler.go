package main

import (
	"fmt"
	"os"
	"time"
)

// allow to retry if there was an error
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

// check if the user input has an error
func checkPrompt(err error) {
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
}
