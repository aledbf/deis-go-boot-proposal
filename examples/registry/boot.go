package main

import (
	"os"

	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	registryCommand = "sudo -E -u registry docker-registry"
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "5000")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/registry")

	bootProcess := boot.New("tcp", externalPort)

	protocol := commons.Getopt("REGISTRY_PROTOCOL", "http")
	secretKey := commons.Getopt("REGISTRY_SECRET_KEY", commons.RandomSSLKey())
	bucketName := commons.Getopt("BUCKET_NAME", "registry")

	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/protocol", protocol)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/secretKey", secretKey)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/bucketName", bucketName)

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-registry...")
	// tune gunicorn
	os.Setenv("GUNICORN_WORKERS", "8")
	bootProcess.StartProcessAsChild(commons.BuildCommandFromString(registryCommand))
	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	hostEtcdPath := commons.Getopt("HOST_ETCD_PATH", "/deis/registry/hosts/"+bootProcess.Host.String())
	bootProcess.Publish(hostEtcdPath, externalPort)
	logger.Log.Info("deis-registry running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-registry...")
	}

	bootProcess.ExecuteOnExit(onExit)
}
