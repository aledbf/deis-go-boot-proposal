set -eo pipefail

main() {
  # initialize database if one doesn't already exist
  # for example, in the case of a data container
  if [[ ! -d /var/lib/postgresql/9.3/main ]]; then
    chown -R postgres:postgres /var/lib/postgresql
    sudo -u postgres /usr/lib/postgresql/9.3/bin/initdb -D /var/lib/postgresql/9.3/main
  fi

  # ensure WAL log bucket exists
  envdir /etc/wal-e.d/env /app/bin/create_bucket ${BUCKET_NAME}

  initial_backup=0
  if [[ ! -f /var/lib/postgresql/9.3/main/initialized ]]; then
    echo "database: no existing database found."
    # check if there are any backups -- if so, let's restore
    # we could probably do better than just testing number of lines -- one line is just a heading, meaning no backups
    if [[ `envdir /etc/wal-e.d/env wal-e --terse backup-list | wc -l` -gt "1" ]]; then
      echo "database: restoring from backup..."
      rm -rf /var/lib/postgresql/9.3/main
      sudo -u postgres envdir /etc/wal-e.d/env wal-e backup-fetch /var/lib/postgresql/9.3/main LATEST
      chown -R postgres:postgres /var/lib/postgresql/9.3/main
      chmod 0700 /var/lib/postgresql/9.3/main
      echo "restore_command = 'envdir /etc/wal-e.d/env wal-e wal-fetch \"%f\" \"%p\"'" | sudo -u postgres tee /var/lib/postgresql/9.3/main/recovery.conf >/dev/null
    else
      echo "database: no backups found. Initializing a new database..."
      initial_backup=1
    fi
    # either way, we mark the database as initialized
    touch /var/lib/postgresql/9.3/main/initialized
  else
    echo "database: existing data directory found. Starting postgres..."
  fi

  # perform a one-time reload to populate database entries
  /usr/local/bin/reload

  if [[ "${initial_backup}" == "1" ]] ; then
    echo "database: performing an initial backup..."
    # perform an initial backup
    sudo -u postgres envdir /etc/wal-e.d/env wal-e backup-push /var/lib/postgresql/9.3/main
  fi
}