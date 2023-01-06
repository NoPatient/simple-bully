package main

import (
	"fmt"
	"os"
	//bully "github.com/NoPatient/simple-bully/"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("you are wrong")
	}

	bully.internal.New

}
