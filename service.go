package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/guobingithub/grpc-load-balance/constants"
	"github.com/guobingithub/grpc-load-balance/logger"
	demopb "github.com/guobingithub/grpc-load-balance/pb"
	"github.com/guobingithub/grpc-load-balance/zkmgr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
)

const (
	Grpc_Addr = "192.168.1.135:%d"
)

var (
	port = flag.Int("p", 50090, "gRPC server port")
)

func main() {
	flag.Parse()
	logger.Info(fmt.Sprintf("my-service start ..."))

	zkConn := registerServer()
	defer func() {
		if zkConn != nil {
			zkConn.Close()
		}
	}()

	StartGrpcServer()
}

type MyDemoService struct {
	ServerAddress string
}

func (ds *MyDemoService) DemoHandler(ctx context.Context, in *demopb.DemoRequest) (*demopb.DemoResponse, error) {
	logger.Error(fmt.Sprintf("MyDemoService, DemoHandler enter, in:(%v)", in))
	name := in.Name
	name = name

	rsp := &demopb.DemoResponse{
		Name: "rspName:" + name + ds.ServerAddress,
	}

	return rsp, nil
}

func registerServer() (conn *zk.Conn) {
	var (
		izk *zkmgr.IZk
		err error
	)

	if izk, err = zkmgr.NewIZk([]string{constants.ZK_Hosts}); err != nil {
		logger.Error(fmt.Sprintf("registerServer, fail to NewIZk. err:%v.", err))
		return
	}

	conn = izk.Conn
	logger.Info(fmt.Sprintf("registerServer, get zk conn:%p.", conn))

	//注册zk节点
	err = izk.RegisterPerServer(constants.ServerName)
	if err != nil {
		logger.Warn(fmt.Sprintf("registerServer, fail to RegisterPerServer node error: %v.", err))
	}

	err = izk.RegisterEphServer(constants.ServerName, constants.RootPath+fmt.Sprintf(Grpc_Addr, *port))
	if err != nil {
		logger.Error(fmt.Sprintf("registerServer, fail to RegisterEphServer node error: %v.", err))
		return
	}

	logger.Info(fmt.Sprintf("registerServer, succeed to RegistServer node."))
	return
}

func registerGRpcServer(s *grpc.Server) {
	demopb.RegisterDemoServiceServer(s, &MyDemoService{ServerAddress: fmt.Sprintf(Grpc_Addr, *port)})
}

func StartGrpcServer() {
	listen, err := net.Listen("tcp", fmt.Sprintf(Grpc_Addr, *port))
	if err != nil {
		grpclog.Fatalf("StartGrpcServer, failed to listen: %v", err)
	}

	// 实例化grpc Server
	s := grpc.NewServer()

	//注册
	registerGRpcServer(s)

	println("StartGrpcServer, Grpc server Listen on ------" + fmt.Sprintf(Grpc_Addr, *port))

	s.Serve(listen)
}
