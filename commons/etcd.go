package commons

import (
	"errors"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/deis/go-boot-proposal/logger"
)

func SetDefaultEtcd(client *etcd.Client, key, value string) {
	SetEtcd(client, key, value, 0)
}

func MkdirEtcd(client *etcd.Client, path string) {
	_, err := client.CreateDir(path, 0)
	if err != nil {
		logger.Log.Debug(err)
	}
}

// WaitForKeysEtcd wait for the required keys up to the timeout or forever if is nil
func WaitForKeysEtcd(client *etcd.Client, keys []string, ttl time.Duration) error {
	start := time.Now()
	wait := true

	for {
		for _, key := range keys {
			_, err := client.Get(key, false, false)
			if err != nil {
				logger.Log.Debugf("key \"%s\" error %v", key, err)
				wait = true
			}
		}

		if !wait {
			return nil
		}

		logger.Log.Debug("waiting for missing etcd keys...")
		time.Sleep(1 * time.Second)
		wait = false

		if time.Since(start) > ttl {
			return errors.New("maximum ttl reached. aborting")
		}
	}
}

func GetEtcd(client *etcd.Client, key string) string {
	result, err := client.Get(key, false, false)
	if err != nil {
		logger.Log.Debugf("%v", err)
		return ""
	}

	return result.Node.Value
}

func GetListEtcd(client *etcd.Client, key string) []string {
	values, err := client.Get(key, true, false)
	if err != nil {
		logger.Log.Debugf("%v", err)
		return []string{}
	}

	result := []string{}
	for _, node := range values.Node.Nodes {
		result = append(result, node.Value)
	}

	logger.Log.Infof("%v", result)
	return result
}

func SetEtcd(client *etcd.Client, key, value string, ttl uint64) {
	_, err := client.Set(key, value, ttl)
	if err != nil {
		logger.Log.Debugf("%v", err)
	}
}

// Publish a service to etcd periodically
func PublishService(
	client *etcd.Client,
	host string,
	etcdPath string,
	externalPort string,
	ttl uint64,
	timeout time.Duration) {

	for {
		SetEtcd(client, etcdPath+"/host", host, ttl)
		SetEtcd(client, etcdPath+"/port", externalPort, ttl)
		time.Sleep(timeout)
	}
}
