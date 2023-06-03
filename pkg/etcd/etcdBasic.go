package etcd

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"mini-k8s/pkg/config"
	"time"
)

type Store struct {
	Client *clientv3.Client
}

func StoreStart() *Store {
	points := make([]string, 1)
	points[0] = "localhost:" + config.EtcdPort1
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   points,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("%v\n", err)
		return &Store{}
	}
	return &Store{
		Client: cli,
	}
}

func (es *Store) StoreEnd() error {
	err := es.Client.Close()
	if err != nil {
		fmt.Printf("%v.\n", err)
		return err
	}
	return nil
}

// EtcdStart used to start a new etcd client instance.
func EtcdStart() *clientv3.Client {
	points := make([]string, 3)
	points[0] = config.IP + ":" + config.EtcdPort1
	points[1] = config.IP + ":" + config.EtcdPort2
	points[2] = config.IP + ":" + config.EtcdPort3
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   points,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("an error occurs when start an etcd client instance: %v\n", err)
		return nil
	}
	return cli
}

// EtcdEnd should be called after all operations finished.
func EtcdEnd(cli *clientv3.Client) error {
	err := cli.Close()
	if err != nil {
		fmt.Printf("an error occurs when stop an etcd instance: %v\n", err)
		return err
	}
	return nil
}
