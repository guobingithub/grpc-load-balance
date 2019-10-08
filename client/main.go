package main

import (
	"context"
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/guobingithub/grpc-load-balance/constants"
	grpcResolver "github.com/guobingithub/grpc-load-balance/grpc-resolve"
	"github.com/guobingithub/grpc-load-balance/logger"
	demopb "github.com/guobingithub/grpc-load-balance/pb"
	"google.golang.org/grpc"
	"strings"
	"time"
)

const (
	RootPath = "/"
)

func main() {
	logger.Error(fmt.Sprintf("grpc-load-balance start..."))

	endpoints := []string{"192.168.1.78:2181"}
	zkConn, _, err := zk.Connect([]string{constants.ZK_Hosts}, 5*time.Second)
	if err != nil {
		logger.Panic(fmt.Sprintf("zk Connect err:(%v)", err))
	}

	resolver := grpcResolver.NewResolver(constants.ServerName, zkConn)
	b := grpc.RoundRobin(resolver)
	conn, err := grpc.DialContext(
		context.TODO(),
		//"/my-service001",//my-service
		strings.Join(endpoints, ","),
		grpc.WithInsecure(),
		grpc.WithBalancer(b),
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Error("DialContext err:(%v)", err)
		return
	}

	t := time.NewTicker(3 * time.Second)
	for tc := range t.C {
		cli := demopb.NewDemoServiceClient(conn)
		res, err := cli.DemoHandler(context.TODO(), &demopb.DemoRequest{Name: "elen"})
		if nil != err {
			panic(err)
		}
		fmt.Println("I has get result ", res.Name, " time is [", tc, "]")
	}

}
