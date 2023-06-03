package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func Del(cli *clientv3.Client, key string) {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := cli.Delete(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("an error occurs when putting data into etcd: %v\n", err)
		return
	}
}

func DelWithPrefix(cli *clientv3.Client, prefix string) {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := cli.Delete(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Printf("an error occurs when putting data into etcd: %v\n", err)
		return
	}
}

func (es *Store) Del(key string) {
	if es.Client == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := es.Client.Delete(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
}

func (es *Store) DelWithPrefix(prefix string) {
	if es.Client == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := es.Client.Delete(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
}
