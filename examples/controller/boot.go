package main

import (
	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	migrationCommand = "sudo -E -u deis /app/manage.py syncdb --migrate --noinput"
	gunicornCommand  = "sudo -E -u deis gunicorn deis.wsgi -b=0.0.0.0 -w=8 -n=deis --timeout=1200 --pid=/tmp/gunicorn.pid --log-level=info --error-logfile=- --access-logfile=-"
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "8000")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/controller")

	bootProcess := boot.New("tcp", externalPort)

	logger.Log.Debug("creating required defaults in etcd...")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/platform")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/services")
	commons.MkdirEtcd(bootProcess.Etcd, "/deis/domains")

	protocol := commons.Getopt("DEIS_PROTOCOL", "http")
	secretKey := commons.Getopt("DEIS_SECRET_KEY", commons.RandomKey())
	builderKey := commons.Getopt("DEIS_BUILDER_KEY", commons.RandomKey())
	registrationEnabled := commons.Getopt("registrationEnabled", "1")
	webEnabled := commons.Getopt("registrationEnabled", "0")

	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/protocol", protocol)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/secretKey", secretKey)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/builderKey", builderKey)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/registrationEnabled", registrationEnabled)
	commons.SetDefaultEtcd(bootProcess.Etcd, etcdPath+"/webEnabled", webEnabled)

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-controller...")

	commons.RunCommand(commons.BuildCommandFromString("mkdir -p /data/logs"))
	commons.RunCommand(commons.BuildCommandFromString("chmod 777 /data/logs"))
	// run an idempotent database migration
	commons.RunCommand(commons.BuildCommandFromString(migrationCommand))
	// run a gunicorn server
	bootProcess.StartProcessAsChild(commons.BuildCommandFromString(gunicornCommand))

	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	hostEtcdPath := commons.Getopt("HOST_ETCD_PATH", "/deis/router/hosts/"+bootProcess.Host.String())

	bootProcess.Publish(hostEtcdPath, externalPort)
	logger.Log.Info("deis-controller running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-controller...")
	}

	bootProcess.ExecuteOnExit(onExit)
}
