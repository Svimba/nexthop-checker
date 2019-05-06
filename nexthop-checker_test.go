package main_test

import (
	"testing"

	nh "github.com/Svimba/nexthop-checker"
)

func TestCheckFlows(t *testing.T) {

	host := "172.16.240.232"
	port := 8085
	progress := false
	var vic nh.VrouterIntrospectCli
	vic.Init(host, port, progress)

	vic.CheckFlows()
	nh.Usage()

	host = "localhost"
	var vic2 nh.VrouterIntrospectCli
	vic2.Init(host, port, progress)
	vic2.CheckFlows()
}
