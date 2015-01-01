package main

import (
	"github.com/progrium/go-basher"
	"github.com/robfig/cron"

	"github.com/deis/go-boot-proposal/boot"
	"github.com/deis/go-boot-proposal/commons"
	"github.com/deis/go-boot-proposal/logger"
)

func main() {
	externalPort := commons.Getopt("EXTERNAL_PORT", "5432")
	etcdPath := commons.Getopt("ETCD_PATH", "/deis/database")

	bootProcess := boot.New("tcp", externalPort)

	adminUser := commons.Getopt("PG_ADMIN_USER", "postgres")
	adminPass := commons.Getopt("PG_ADMIN_PASS", "changeme123")
	user := commons.Getopt("PG_USER_NAME", "deis")
	password := commons.Getopt("PG_USER_PASS", "changeme123")
	name := commons.Getopt("PG_USER_DB", "deis")
	bucketName := commons.Getopt("BUCKET_NAME", "db_wal")
	backupsToRetain := commons.Getopt("BACKUPS_TO_RETAIN", "5")
	backupFrequency := commons.Getopt("BACKUP_FREQUENCY", "3h")
	pgConfig := commons.Getopt("PG_CONFIG", "/etc/postgresql/9.3/main/postgresql.conf")
	listenAddress := commons.Getopt("PG_LISTEN", "*")

	logger.Log.Debug("creating required defaults in etcd...")
	commons.SetDefaultEtcd(bootProcess.Etcd, "engine", "postgresql_psycopg2")
	commons.SetDefaultEtcd(bootProcess.Etcd, "adminUser", adminUser)
	commons.SetDefaultEtcd(bootProcess.Etcd, "adminPass", adminPass)
	commons.SetDefaultEtcd(bootProcess.Etcd, "user", user)
	commons.SetDefaultEtcd(bootProcess.Etcd, "password", password)
	commons.SetDefaultEtcd(bootProcess.Etcd, "name", name)
	commons.SetDefaultEtcd(bootProcess.Etcd, "bucketName", bucketName)

	bash, _ := basher.NewContext("/bin/bash", false)
	bash.Source("bash/postgres.bash", nil)
	bash.Export("BUCKET_NAME", bucketName)
	_, err := bash.Run("main", nil)
	if err != nil {
		logger.Log.Fatal(err)
	}

	startedChan := make(chan bool)
	logger.Log.Info("starting deis-database...")
	bootProcess.StartProcessAsChild("sudo", "-i", "-u", "postgres",
		"/usr/lib/postgresql/9.3/bin/postgres",
		"-c", "config-file="+pgConfig,
		"-c", "listen-addresses="+listenAddress)
	bootProcess.WaitForLocalConnection(startedChan)
	<-startedChan

	bootProcess.Publish(etcdPath, externalPort)
	logger.Log.Info("deis-database running...")

	// schedule periodic backups using wal-e
	scheduleBackup := cron.New()
	scheduleBackup.AddFunc("@every "+backupFrequency, func() { backupPostgres(backupsToRetain) })
	scheduleBackup.Start()

	onExit := func() {
		logger.Log.Debug("terminating deis-database...")
	}

	bootProcess.ExecuteOnExit(onExit)
}

func backupPostgres(backupsToRetain string) {
	bash, _ := basher.NewContext("/bin/bash", false)
	bash.Source("bash/backup.bash", nil)
	bash.Export("BACKUPS_TO_RETAIN", backupsToRetain)
	_, err := bash.Run("main", nil)
	if err != nil {
		logger.Log.Fatal(err)
	}
}
