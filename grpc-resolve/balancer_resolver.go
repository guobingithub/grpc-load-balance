package grpcResolver

import (
	"errors"

	"github.com/go-zookeeper/zk"
	log "github.com/guobingithub/grpc-load-balance/logger"
	"google.golang.org/grpc/naming"
)

// 实现naming Resolver接口
type resolver struct {
	serviceName string   //服务名称 对应zk中的key
	zkCli       *zk.Conn //zk client.
	w           naming.Watcher
}

func NewResolver(serviceName string, cli *zk.Conn) *resolver {
	r := new(resolver)
	r.serviceName = serviceName
	r.zkCli = cli
	r.w = &watcher{resolver: r}
	return r
}

//Resolve 根据服务名称生成watcher.
func (r *resolver) Resolve(target string) (naming.Watcher, error) {
	log.Debug("resolver.Resolve()")
	if r.serviceName == "" {
		return nil, errors.New("grpclb: no service name provided")
	}
	return r.w, nil
}
