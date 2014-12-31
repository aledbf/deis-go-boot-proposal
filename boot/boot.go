package boot

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-etcd/etcd"

	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	timeout time.Duration = 10 * time.Second
	ttl     time.Duration = timeout * 2
	wait    time.Duration = timeout / 2
)

func (boot *Boot) Prepare(componentName string) {
}

func New(protocol string, port string) *Boot {
	logger.Log.Info("Starting deis component...")

	host := commons.Getopt("HOST", "127.0.0.1")

	etcdPort := commons.Getopt("ETCD_PORT", "4001")

	etcdClient := etcd.NewClient([]string{"http://" + host + ":" + etcdPort})

	// wait until etcd has discarded potentially stale values
	time.Sleep(timeout + 1)

	etcdHostPort := host + ":" + etcdPort

	// wait for confd to run once and install initial templates
	commons.WaitForInitialConfd(etcdHostPort, timeout)

	// spawn confd in the background to update services based on etcd changes
	commons.LaunchConfd(etcdHostPort)

	return &Boot{
		Etcd:     etcdClient,
		Confd:    "",
		Host:     net.ParseIP(host),
		Timeout:  timeout,
		TTL:      timeout * 2,
		Protocol: protocol,
		Port:     port,
	}
}

func (this *Boot) WaitForInitialConfd() {
}

func (this *Boot) Publish(path string, port string) {
	go commons.PublishService(this.Etcd, this.Host.String(), path, port, uint64(ttl.Seconds()), timeout)
}

func (this *Boot) StartProcessAsChild(command string, args ...string) {
	go commons.StartServiceCommand(command, args...)
}

func (this *Boot) WaitForLocalConnection(startedChan chan bool, port string) {
	if port == "" {
		port = this.Port
	}

	go commons.WaitForLocalConnection(startedChan, this.Protocol, port)
}

// ExecuteOnExit tasks to be executed when the process ends (included ctrl+c)
func (this *Boot) ExecuteOnExit(functions ...func()) {
	exitChan := make(chan os.Signal, 2)
	signal.Notify(exitChan, syscall.SIGTERM, syscall.SIGINT)
	for _, function := range functions {
		function()
	}
	<-exitChan
}
