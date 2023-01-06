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
		"0": "0.0.0.0:10001",
		"1": "0.0.0.0:10002",
		"2": "0.0.0.0:10003",
		"3": "0.0.0.0:10004",
		"4": "0.0.0.0:10005",
	})

	for _, k := range viper.AllKeys() {
		log.Printf("%s, %+v", k, viper.AllSettings()[k])
	}
}
