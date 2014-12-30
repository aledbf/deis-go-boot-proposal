package boot

import (
	"net"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

type Boot struct {
	Etcd     *etcd.Client
	Confd    string
	Host     net.IP
	Timeout  time.Duration
	TTL      time.Duration
	Port     string
	Protocol string
}
