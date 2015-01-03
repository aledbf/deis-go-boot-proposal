package main

import (
	"os"

	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/ceph"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	etcdPath = "/deis/store/osds"
)

func main() {
	bootProcess := boot.New("tcp", "6800")

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-store-daemon...")

	commons.MkdirEtcd(bootProcess.Etcd, etcdPath)

	var osdID string
	if !ceph.IsOSDCreated(bootProcess.Etcd) {
		logger.Log.Info("creating OSD...")
		osdID = ceph.CreateOSD(bootProcess.Etcd)
	} else {
		logger.Log.Debug("OSD already created...")
		osdID = ceph.GetOSDID(bootProcess.Etcd)
	}

	// Make sure osd directory exists
	os.MkdirAll("/var/lib/ceph/osd/ceph-"+osdID, 0755)

	if !ceph.IsOSDInitialized(osdID) {
		logger.Log.Info("OSD not yet initialized. Initializing...")
		commons.RunCommand(commons.BuildCommandFromString("ceph-osd -i " + osdID + " --mkfs --mkjournal --osd-journal /var/lib/ceph/osd/ceph-" + osdID + "/journal"))
		commons.RunCommand(commons.BuildCommandFromString("ceph auth get-or-create osd." + osdID + " osd 'allow *' mon 'allow profile osd' -o /var/lib/ceph/osd/ceph-" + osdID + "/keyring"))
		hostname := commons.RunCommand(commons.BuildCommandFromString("hostname"))
		commons.RunCommand(commons.BuildCommandFromString("ceph osd crush add " + osdID + " 1.0 root=default host=" + hostname))
	}

	bootProcess.StartProcessAsChild(commons.BuildCommandFromString("ceph-osd -d -i " + osdID + " -k /var/lib/ceph/osd/ceph-" + osdID + "/keyring"))
	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	bootProcess.Publish(etcdPath+"/"+bootProcess.Host.String(), "6800")
	logger.Log.Info("deis-store-daemon running...")

	onExit := func() {
		logger.Log.Debug("terminating deis-store-daemon...")
	}

	bootProcess.ExecuteOnExit(onExit)
}
