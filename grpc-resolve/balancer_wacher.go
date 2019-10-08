package grpcResolver

import (
	"github.com/go-zookeeper/zk"
	log "github.com/guobingithub/grpc-load-balance/logger"
	"google.golang.org/grpc/naming"
)

// 实现naming Watcher 接口
type watcher struct {
	resolver      *resolver //zk Resolver
	isInitialized bool
}

// 关闭watcher
func (watcher *watcher) Close() {
	watcher.isInitialized = false
	log.Info("watcher.Close()")
}

// 监听特定值
func (watcher *watcher) Next() ([]*naming.Update, error) {
	// 是否初始化
	//log.Info("watcher Next, %p", watcher)
	serviceName := watcher.resolver.serviceName
	if !watcher.isInitialized {
		// 获取服务端地址
		// "/gomicro-server"
		//list, _, err := watcher.resolver.zkCli.Children("/"+serviceName)
		list, _, err := watcher.resolver.zkCli.Children(serviceName)
		watcher.isInitialized = true
		if err == nil {
			addrs := list
			if l := len(addrs); l != 0 {
				updates := make([]*naming.Update, l)
				for i := range addrs {
					updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
					log.Info("watcher next initial new addr:%s, srvName:%s, %p", addrs[i], serviceName, watcher)
				}
				return updates, nil
			}
		}
	}
	// 监听
	//snapshot, _, events, err := watcher.resolver.zkCli.ChildrenW("/"+serviceName)
	snapshot, _, events, err := watcher.resolver.zkCli.ChildrenW(serviceName)
	if err == nil {
		return nil, nil
	}
	for i := range snapshot {
		for ev := range events {
			log.Info("watcher watch event key:%s, value:%s, srvName:%s, type:%d, %p",
				string(ev.Path), string(ev.Type), serviceName, ev.Type, watcher)
			switch ev.Type {
			case zk.EventNodeCreated:
				return []*naming.Update{{Op: naming.Add, Addr: snapshot[i]}}, nil
			case zk.EventNodeDeleted:
				return []*naming.Update{{Op: naming.Delete, Addr: snapshot[i]}}, nil
			}
		}
	}
	return nil, nil
}
