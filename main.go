package main

import (
	"fmt"
	"os"
	"strconv"
	"github.com/AndochBonin/ttype/tui"
)

func main() {
	testDurationSeconds := 30
	if len(os.Args) > 2 {
		fmt.Println("too many args")
		return
	}
	if len(os.Args) == 2 {
		if duration, err := strconv.Atoi(os.Args[1]); err != nil {
			fmt.Println("argument is not valid duration")
			return
		} else {
			testDurationSeconds = duration
		}
	}
	err := tui.Run(testDurationSeconds)
	if err != nil {
		fmt.Println("boohoo")
	}
}