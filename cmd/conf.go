package main

import (
	"github.com/spf13/viper"
	"log"
)

const (
	//confFile     = "bully.conf"
	confDataAddr = "data_server_address"
	confPeerAddr = "peer_address"
)

func init() {
	viper.SetDefault(confDataAddr, "0.0.0.8081")
	viper.SetDefault(confPeerAddr, map[string]string{
		"0": "127.0.0.1:10001",
		"1": "127.0.0.1:10002",
		"2": "127.0.0.1:10003",
		"3": "127.0.0.1:10004",
		"4": "127.0.0.1:10005",
	})

	for _, k := range viper.AllKeys() {
		log.Printf("%s, %+v", k, viper.AllSettings()[k])
	}
}
