package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"time"
)

// Watch change API return type here(lyh).
func Watch(cli *clientv3.Client, key string) {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	rch := cli.Watch(ctx, key)
	cancel()
	for resp := range rch {
		for _, ev := range resp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

func WatchNew(cli *clientv3.Client, key string) clientv3.WatchChan {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when watching data in etcd!")
		return nil
	}
	rch := cli.Watch(context.Background(), key)
	return rch
}

func WatchWithPrefix(cli *clientv3.Client, prefix string) clientv3.WatchChan {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when watching prefix in etcd!")
		return nil
	}
	rch := cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	return rch
}

func (es *Store) WatchWithPrefix(prefix string) clientv3.WatchChan {
	return es.Client.Watch(context.Background(), prefix, clientv3.WithPrefix())
}
