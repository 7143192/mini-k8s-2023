package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func Get(cli *clientv3.Client, key string) *clientv3.GetResponse {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	response, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("an error occurs when putting data into etcd: %v\n", err)
		return nil
	}
	return response
}

func GetWithPrefix(cli *clientv3.Client, prefix string) *clientv3.GetResponse {
	if cli == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	response, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Printf("an error occurs when putting data into etcd: %v\n", err)
		return nil
	}
	return response
}

func GetPodsChanges(cur []string, old []string) ([]string, []string) {
	fmt.Printf("cur set in GetPodsChanges = %v\n", cur)
	fmt.Printf("old set in GetPodsChanges = %v\n", old)
	newAdd := make([]string, 0)
	newDel := make([]string, 0)
	for _, curStr := range cur {
		found := false
		for _, oldStr := range old {
			if oldStr == curStr {
				found = true
			}
		}
		if found == false {
			// cur has but old doesn't have, it is a new added pod instance.
			newAdd = append(newAdd, curStr)
		}
	}
	for _, oldStr := range old {
		found := false
		for _, curStr := range cur {
			if curStr == oldStr {
				found = true
			}
		}
		if found == false {
			// old has but cur doesn't have, it is a deleted pod instance.
			newDel = append(newDel, oldStr)
		}
	}
	return newAdd, newDel
}

func (es *Store) Get(key string) *clientv3.GetResponse {
	if es.Client == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	response, err := es.Client.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	return response
}

func (es *Store) GetWithPrefix(prefix string) *clientv3.GetResponse {
	if es.Client == nil {
		fmt.Println("client-v3 is not initialized when putting data into etcd!")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	response, err := es.Client.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	return response
}
