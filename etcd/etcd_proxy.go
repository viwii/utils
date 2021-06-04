package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// EtcdProxy 管理工具
type EtcdProxy struct {
	client *clientv3.Client
}

var ETCD_TRANSPORT_TIMEOUT = 5 * time.Second

var _etcdProxy *EtcdProxy

// Init 链接
func (ep *EtcdProxy) Init(ipStr string) {
	var err error
	ep.client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{ipStr},
		DialTimeout: time.Duration(30) * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// DownloadAndWatch 下载及观察文件
func (ep *EtcdProxy) DownloadAndWatch(path string, Callback func(data string, err error)) {
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_TRANSPORT_TIMEOUT)
	resp, err := ep.client.Get(ctx, path)
	cancel()

	if err == nil {
		for _, ev := range resp.Kvs {
			Callback(string(ev.Key), nil)
		}

		go func() {
			rch := ep.client.Watch(context.Background(), path)
			for wresp := range rch {
				for _, ev := range wresp.Events {
					log.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					if ev.Type == 0 {
						Callback(string(ev.Kv.Key), nil)
					}
				}
			}
		}()
	}

	Callback("", err)
}

func (ep *EtcdProxy) RegisterServe(appName string, serverInfo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ETCD_TRANSPORT_TIMEOUT)
	leaseResp, err := ep.client.Grant(ctx, 10) //租约时间设定为10秒
	cancel()
	if err != nil {
		return err
	}

	_, err = ep.client.Put(context.TODO(), appName, serverInfo, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		log.Fatal(err)
	}

	// the key 'foo' will be kept forever
	ch, kaerr := ep.client.KeepAlive(context.TODO(), leaseResp.ID)
	if kaerr != nil {
		log.Fatal(kaerr)
	}

	ka := <-ch
	fmt.Println("ttl:", ka.TTL)

	return nil
}

// NewEtcdProxy Etcd管理代理
func NewEtcdProxy(ipStr string) *EtcdProxy {
	if _etcdProxy == nil {
		_etcdProxy = new(EtcdProxy)
		_etcdProxy.Init(ipStr)
	}

	return _etcdProxy
}
