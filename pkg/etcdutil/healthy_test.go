package etcdutil

import (
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func TestHeartbeat(t *testing.T) {
	name := "TestHeartbeat"
	m := StartNewEtcdServer(t, name)
	defer m.Terminate(t)

	client := etcd.NewClient([]string{m.URL()})
	taskID := uint64(0)
	ttl := uint64(1)
	interval := time.Duration(ttl) * time.Second
	stop := make(chan struct{}, 1)

	client.Create(HealthyPath(name, taskID), "health", ttl)
	time.Sleep(2 * interval)
	_, err := client.Get(HealthyPath(name, taskID), false, false)
	if err == nil {
		t.Fatal("ttl node should expire")
	}

	client.Create(HealthyPath(name, taskID), "health", ttl)
	go Heartbeat(client, name, taskID, interval, stop)
	time.Sleep(10 * interval)
	_, err = client.Get(HealthyPath(name, taskID), false, false)
	if err != nil {
		t.Fatalf("client.Get failed: %v", err)
	}

	close(stop)
	time.Sleep(10 * interval)
	_, err = client.Get(HealthyPath(name, taskID), false, false)
	if err == nil {
		t.Fatal("ttl node should expire")
	}
}
