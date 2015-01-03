package main

import (
	"os"
	"syscall"

	"github.com/ActiveState/tail"
	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	nginxCommand          = "/opt/nginx/sbin/nginx -c /opt/nginx/conf/nginx.conf"
	gitLogFile     string = "/opt/nginx/logs/git.log"
	nginxAccessLog string = "/opt/nginx/logs/access.log"
	nginxErrorLog  string = "/opt/nginx/logs/error.log"
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "80")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/router")

	bootProcess := boot.New("tcp", externalPort)

	logger.Log.Debug("creating required defaults in etcd...")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/controller")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/services")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/domains")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/builder")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/router/hosts")

	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/gzip", "on")

	// tail the logs
	go tailFile(nginxAccessLog)
	go tailFile(nginxErrorLog)
	go tailFile(gitLogFile)

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-router...")
	bootProcess.StartProcessAsChild(commons.BuildCommandFromString(nginxCommand))
	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	hostEtcdPath := commons.Getopt("HOST_ETCD_PATH", "/deis/router/hosts/"+bootProcess.Host.String())

	bootProcess.Publish(hostEtcdPath, externalPort)
	logger.Log.Info("deis-router running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-router...")
		tail.Cleanup()
	}

	bootProcess.ExecuteOnExit(onExit)
}

func tailFile(path string) {
	mkfifo(path)
	t, _ := tail.TailFile(path, tail.Config{Follow: true})

	for line := range t.Lines {
		logger.Log.Info(line.Text)
	}
}

func mkfifo(path string) {
	os.Remove(path)
	if err := syscall.Mkfifo(path, syscall.S_IFIFO|0666); err != nil {
		logger.Log.Fatalf("%v", err)
	}
}
