package ceph

import (
	"os"
	"regexp"

	"github.com/coreos/go-etcd/etcd"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

const (
	osdIDRegex string = `^-?[0-9]+$`
)

// IsOSDCreated check if there is a OSD for localhost
func IsOSDCreated(client *etcd.Client) bool {
	host := commons.Getopt("HOST", "")
	if _, err := client.Get("/deis/store/osds/"+host, false, false); err != nil {
		return false
	}

	return true
}

func GetOSDID(client *etcd.Client) string {
	host := commons.Getopt("HOST", "")
	osdId := commons.GetEtcd(client, "/deis/store/osds/"+host)
	checkOSDId(osdId)
	return osdId
}

// CreateOSD executes ceph osd create and returns the ID
func CreateOSD(client *etcd.Client) string {
	output := commons.RunCommand("ceph", []string{"osd", "create"})
	checkOSDId(output)
	logger.Log.Info("OSD_ID: " + output)
	return output
}

// IsOSDInitialized checks if the OSD with ID osdID exists in the local filesystem
func IsOSDInitialized(osdID string) bool {
	_, err := os.Stat("/var/lib/ceph/osd/ceph-" + osdID + "/keyring")
	return !os.IsNotExist(err)
}

func checkOSDId(osdId string) {
	r := regexp.MustCompile(osdIDRegex)
	match := r.FindStringSubmatch(osdId)
	if match == nil {
		logger.Log.Error("We have an OSD ID that isn't an integer")
		logger.Log.Error("This likely means the monitor we tried to connect to isn't up, but others may be.")
		logger.Log.Fatal("We can't proceed because we don't know if an OSD was created or not.")
	}
}
