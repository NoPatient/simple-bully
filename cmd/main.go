package main

import (
	"fmt"
	bully "github.com/NoPatient/simple-bully"
	"github.com/spf13/viper"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("you are wrong")
	}
	configNodes := viper.GetStringMapString(confPeerAddr)
	b, _ := bully.NewBully("0", "127.0.0.1:10001", "tcp4", configNodes)
	workFunc := func() {
		for {
			fmt.Printf("Bully %s: Coordinator is %s\n", b.ID, b.GetCoordinator())
			time.Sleep(1 * time.Second)
		}
	}

	b.Run(workFunc)

}
