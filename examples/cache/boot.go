package main

import (
	"io/ioutil"
	"strings"

	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	redisConf     string = "/app/redis.conf"
	defaultMemory string = "50mb"
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "6379")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/cache")

	bootProcess := boot.New("tcp", externalPort)

	maxmemory := commons.GetEtcd(bootProcess.Etcd, "/deis/cache/maxmemory")
	if maxmemory == "" {
		maxmemory = defaultMemory
	}
	replaceMaxmemoryInConfig(maxmemory)

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-cache...")
	bootProcess.StartProcessAsChild("/app/bin/redis-server", redisConf)
	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	bootProcess.Publish(etcdPath, externalPort)
	logger.Log.Info("deis-cache running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-cache...")
	}

	bootProcess.ExecuteOnExit(onExit)
}

func replaceMaxmemoryInConfig(maxmemory string) {
	input, err := ioutil.ReadFile(redisConf)
	if err != nil {
		logger.Log.Fatalln(err)
	}
	output := strings.Replace(string(input), "# maxmemory <bytes>", "maxmemory "+maxmemory, 1)
	err = ioutil.WriteFile(redisConf, []byte(output), 0644)
	if err != nil {
		logger.Log.Fatalln(err)
	}
}
