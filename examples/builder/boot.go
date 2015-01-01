package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-etcd/etcd"

	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/builder"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

func main() {
	logger.Log.Info("starting deis-builder...")

	etcdPath := commons.Getopt("ETCD_PATH", "/deis/builder")
	externalPort := commons.Getopt("EXTERNAL_PORT", "2223")

	bootProcess := boot.New("tcp", externalPort)

	storageDriver := commons.Getopt("STORAGE_DRIVER", "btrfs")

	commons.MkdirEtcd(bootProcess.Etcd, etcdPath)
	commons.MkdirEtcd(bootProcess.Etcd, etcdPath+"/users")

	// wait until etcd has discarded potentially stale values
	time.Sleep(bootProcess.Timeout + 1)

	// check for stored configuration in deis-store
	builder.CheckSSHKeysInStore(bootProcess.Etcd)

	// remove any pre-existing docker.sock
	// spawn a docker daemon to run builds
	os.Remove("/var/run/docker.sock")

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-builder...")

	go bootProcess.StartProcessAsChild("docker", "-d", "--storage-driver="+storageDriver, "--bip=172.19.42.1/16")

	// wait for docker to start
	waitForDocker()

	// HACK: load cedarish, slugbuilder and slugrunner from the local registry
	checkCedarish(bootProcess.Etcd)

	logger.Log.Debug("starting ssh server...")
	// start an SSH daemon to process `git push` requests
	bootProcess.StartProcessAsChild("/usr/sbin/sshd", "-D", "-e", "-E", "/app/ssh.log")
	bootProcess.WaitForLocalConnection(startedChan, "22")
	<-startedChan

	bootProcess.Publish(etcdPath, externalPort)

	// Wait for terminating signal
	onExit := func() {
		logger.Log.Debug("terminating deis-builder...")
	}

	bootProcess.ExecuteOnExit(onExit)
}

func waitForDocker() {
	logger.Log.Debug("waiting for docker daemon to be available...")
	for {
		cmd := exec.Command("docker", "info")
		if err := cmd.Run(); err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}
}

func checkCedarish(client *etcd.Client) {
	logger.Log.Debug("checking for cedarish (slugrunner/slugbuilder)...")

	slugrunner := commons.GetEtcd(client, "/deis/slugrunner/image")
	logger.Log.Debugf("checking if image %s exists locally", slugrunner)
	docker, _ := commons.NewDockerClient()
	_, err := docker.InspectImage(slugrunner)
	if err != nil {
		logger.Log.Warn("slugrunner is missing. building a new one...")
		cmd := exec.Command("docker", "pull", slugrunner)
		cmd.Run()

		cmd = exec.Command("docker", "tag", slugrunner, "deis/slugrunner")
		cmd.Run()
	} else {
		logger.Log.Info("slugrunner already loaded")
	}

	slugbuilder := commons.GetEtcd(client, "/deis/slugbuilder/image")
	logger.Log.Debugf("checking if image %s exists locally", slugbuilder)
	docker, _ = commons.NewDockerClient()
	_, err = docker.InspectImage(slugbuilder)
	if err != nil {
		logger.Log.Warn("slugbuilder is missing. building a new one")
		cmd := exec.Command("docker", "pull", slugbuilder)
		cmd.Run()

		cmd = exec.Command("docker", "tag", slugbuilder, "deis/slugbuilder")
		cmd.Run()
	} else {
		logger.Log.Info("slugbuilder already loaded")
	}
}
