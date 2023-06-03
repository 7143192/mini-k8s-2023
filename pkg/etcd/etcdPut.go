package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func Put(cli *clientv3.Client, key string, val string) {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := cli.Put(ctx, key, val)
	cancel()
	if err != nil {
		fmt.Printf("an error occurs when putting data into etcd: %v\n", err)
		return
	}
}

func (es *Store) Put(key, val string) {
	if es.Client == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := es.Client.Put(ctx, key, val)
	cancel()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
}
