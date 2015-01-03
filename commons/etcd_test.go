package commons

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/deis/go-boot-proposal/logger"
)

func init() {
	_, err := exec.Command("etcd", "--version").Output()
	if err != nil {
		logger.Log.Fatal(err)
	}
}

var etcdServer *exec.Cmd

func startEtcd() {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "etcd-test")
	if err != nil {
		logger.Log.Fatal("creating temp dir:", err)
	}
	logger.Log.Debugf("temp dir: %v", tmpDir)

	etcdServer = exec.Command("etcd", "-data-dir="+tmpDir, "-name=default")
	etcdServer.Start()
	time.Sleep(1 * time.Second)
}

func stopEtcd() {
	etcdServer.Process.Kill()
}

func TestGetSetEtcd(t *testing.T) {
	startEtcd()
	defer stopEtcd()

	etcdClient := etcd.NewClient([]string{"http://localhost:4001"})
	SetDefaultEtcd(etcdClient, "/path", "value")
	value := GetEtcd(etcdClient, "/path")

	if value != "value" {
		t.Fatalf("Expected '%v' arguments but returned '%v'", "value", value)
	}

	SetDefaultEtcd(etcdClient, "/path", "")
	value = GetEtcd(etcdClient, "/path")

	if value != "" {
		t.Fatalf("Expected '%v' arguments but returned '%v'", "", value)
	}

	SetEtcd(etcdClient, "/path", "value", uint64((1 * time.Second).Seconds()))
	time.Sleep(1200 * time.Millisecond)
	value = GetEtcd(etcdClient, "/path")

	if value != "" {
		t.Fatalf("Expected '%v' arguments but returned '%v'", "", value)
	}
}

func TestMkdirEtcd(t *testing.T) {
	startEtcd()
	defer stopEtcd()

	etcdClient := etcd.NewClient([]string{"http://localhost:4001"})

	MkdirEtcd(etcdClient, "/directory")
	values := GetListEtcd(etcdClient, "/directory")
	if len(values) != 0 {
		t.Fatalf("Expected '%v' arguments but returned '%v'", 0, len(values))
	}

	SetEtcd(etcdClient, "/directory/item_1", "value", 0)
	SetEtcd(etcdClient, "/directory/item_2", "value", 0)
	values = GetListEtcd(etcdClient, "/directory")
	if len(values) != 2 {
		t.Fatalf("Expected '%v' arguments but returned '%v'", 2, len(values))
	}

}
