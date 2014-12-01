package integration

import (
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/go-distributed/meritop/pkg/etcdutil"
)

func TestHeartbeat(t *testing.T) {
	name := "TestHeartbeat"
	m := etcdutil.StartNewEtcdServer(t, name)
	defer m.Terminate(t)

	client := etcd.NewClient([]string{m.URL()})
	taskID := uint64(1)
	ttl := uint64(1)
	interval := time.Duration(ttl) * time.Second
	stop := make(chan struct{}, 1)

	client.Create(etcdutil.HealthyPath(name, taskID), "health", ttl)
	time.Sleep(2 * interval)
	_, err := client.Get(etcdutil.HealthyPath(name, taskID), false, false)
	if err == nil {
		t.Fatal("ttl node should expire")
	}

	client.Create(etcdutil.HealthyPath(name, taskID), "health", ttl)
	go etcdutil.Heartbeat(client, name, taskID, interval, stop)
	time.Sleep(10 * interval)
	_, err = client.Get(etcdutil.HealthyPath(name, taskID), false, false)
	if err != nil {
		t.Fatalf("client.Get failed: %v", err)
	}

	close(stop)
	time.Sleep(10 * interval)
	_, err = client.Get(etcdutil.HealthyPath(name, taskID), false, false)
	if err == nil {
		t.Fatal("ttl node should expire")
	}
}

func TestDetectFailure(t *testing.T) {
	name := "TestDetectFailure"
	m := etcdutil.StartNewEtcdServer(t, name)
	defer m.Terminate(t)

	client := etcd.NewClient([]string{m.URL()})
	taskID := uint64(1)
	ttl := uint64(1)
	failure := make(chan uint64, 1)

	client.Create(etcdutil.HealthyPath(name, taskID), "health", ttl)
	go func() {
		failure <- etcdutil.DetectFailure(client, name, taskID, nil)
	}()
	failedTaskID := <-failure
	if failedTaskID != taskID {
		t.Fatalf("failedTaskID want = %d, get = %d", taskID, failedTaskID)
	}
}
