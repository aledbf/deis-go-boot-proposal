package main

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ActiveState/tail"
	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	gitLogFile     string        = "/opt/nginx/logs/git.log"
	nginxAccessLog string        = "/opt/nginx/logs/access.log"
	nginxErrorLog  string        = "/opt/nginx/logs/error.log"
	timeout        time.Duration = 10 * time.Second
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "80")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/router")

	process := boot.New("tcp", externalPort)

	logger.Log.Debug("creating required defaults in etcd...")
	commons.MkdirEtcd(process.Etcd, "/deis/controller")
	commons.MkdirEtcd(process.Etcd, "/deis/services")
	commons.MkdirEtcd(process.Etcd, "/deis/domains")
	commons.MkdirEtcd(process.Etcd, "/deis/builder")
	commons.MkdirEtcd(process.Etcd, "/deis/router/hosts")

	commons.SetDefaultEtcd(process.Etcd, etcdPath+"/gzip", "on")

	// tail the logs
	go tailFile(nginxAccessLog)
	go tailFile(nginxErrorLog)
	go tailFile(gitLogFile)

	nginxChan := make(chan bool)
	logger.Log.Info("starting deis-router...")
	go launchNginx(nginxChan, "tcp", externalPort, "/opt/nginx/sbin/nginx", "-c", "/opt/nginx/conf/nginx.conf")
	<-nginxChan

	hostEtcdPath := commons.Getopt("HOST_ETCD_PATH", "/deis/router/hosts/"+process.Host.String())

	init.Publish(hostEtcdPath, externalPort)
	logger.Log.Info("deis-router running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-router...")
		tail.Cleanup()
	}

	init.ExecuteOnExit(onExit)
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
