// MIT License
//
// (C) Copyright [2019-2023] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const APP_VERSION = "1"
const SCHEMA_VERSION = 20
const SCHEMA_STEPS = 22

var dbName string
var dbUser string
var dbPass string
var dbHost string
var dbPortStr string
var dbPort int
var dbOpts string
var forceStep int
var migrationsDir string
var migrateStep uint
var fresh *bool
var versionFlag *bool

// Parse command line options.
func parseCmdLine() {
	flag.StringVar(&dbName, "dbname", "", "Database name (default 'hmsds'")
	flag.StringVar(&dbUser, "dbuser", "", "Database user name")
	flag.StringVar(&dbHost, "dbhost", "", "Database hostname")
	flag.StringVar(&dbPortStr, "dbport", "", "Database port")
	flag.StringVar(&dbOpts, "dbopts", "", "Database options string")
	flag.IntVar(&forceStep, "f", -1, "Force migration to step X")
	flag.StringVar(&migrationsDir, "migrationsdir", "/persistent_migrations", "Directory with migrations v4 files")
	fresh = flag.Bool("fresh", false,
		"Revert all schemas before installing (drops all data)")
	versionFlag = flag.Bool("v", false, "Print the version number.")

	flag.Parse()

	if *versionFlag {
		lg.Printf("Version: %s, Schema version: %d",
			APP_VERSION, SCHEMA_VERSION)
	}
	envvar := "SMD_DBNAME"
	if dbName == "" {
		if val := os.Getenv(envvar); val != "" {
			dbName = val
		}
	}
	envvar = "SMD_DBUSER"
	if dbUser == "" {
		if val := os.Getenv(envvar); val != "" {
			dbUser = val
		}
	}
	envvar = "SMD_DBHOST"
	if dbHost == "" {
		if val := os.Getenv(envvar); val != "" {
			dbHost = val
		}
	}
	envvar = "SMD_DBPORT"
	if dbPortStr == "" {
		if val := os.Getenv(envvar); val != "" {
			dbPortStr = val
		}
	}
	if dbPortStr == "" {
		lg.Printf("Missing DB port number")
		flag.Usage()
		os.Exit(1)
	} else {
		port, err := strconv.ParseInt(dbPortStr, 10, 64)
		if err != nil {
			lg.Printf("Bad dbport '%s': %s", dbPortStr, err)
			flag.Usage()
			os.Exit(1)
		}
		dbPort = int(port)
	}
	envvar = "SMD_DBOPTS"
	if dbOpts == "" {
		if val := os.Getenv(envvar); val != "" {
			dbOpts = val
		}
	}
	envvar = "SMD_FORCESTEP"
	if forceStep == -1 {
		if val := os.Getenv(envvar); val != "" {
			step, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				lg.Printf("Bad step for forceStep: '%s': %s", val, err)
				flag.Usage()
				os.Exit(1)
			}
			forceStep = int(step)
			if forceStep < 0 || forceStep > SCHEMA_STEPS {
				lg.Printf("invalid step for forceStep: '%s': out of range", val)
				flag.Usage()
				os.Exit(1)
			}
		}
	}
	envvar = "SMD_DBSTEPS"
	if val := os.Getenv(envvar); val != "" {
		step, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			lg.Printf("Bad step for DB Steps: '%s': %s", val, err)
			os.Exit(1)
		}
		if step < 0 || step > SCHEMA_STEPS {
			lg.Printf("invalid step for DB Steps: '%s': out of range.", val)
			os.Exit(1)
		}
		migrateStep = uint(step)
	} else {
		migrateStep = uint(SCHEMA_STEPS)
	}
	envvar = "SMD_FRESH"
	if !*fresh {
		if val := os.Getenv(envvar); val != "" {
			*fresh = true
		}
	}

	// Env var only
	envvar = "SMD_DBPASS"
	if val := os.Getenv(envvar); val != "" {
		dbPass = val
	}

	// Set default dbName
	if dbName == "" {
		dbName = "hmsds"
	}
	// Set default dbUser
	if dbUser == "" {
		dbUser = "hmsdsuser"
	}
}

var lg = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)

func main() {
	PrintVersionInfo()
	parseCmdLine()

	lg.Printf("smd-init: Starting...")
	lg.Printf("smd-init: Version: %s, SchemaVersion: %d, Steps: %d, Desired Step: %d",
		APP_VERSION, SCHEMA_VERSION, SCHEMA_STEPS, migrateStep)

	dbDSN := hmsds.GenDsnHMSDB_PB(dbName, dbUser, dbPass, dbHost, dbOpts, dbPort)
	if dbDSN == "" {
		lg.Printf("Empty DSN created via flag or db options\n")
		flag.Usage()
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dbDSN)
	for {
		if err != nil {
			lg.Printf("Open failed: '%s'", err)
			lg.Printf("Retrying after 5 seconds...")
			time.Sleep(5 * time.Second)
			db, err = sql.Open("postgres", dbDSN)
		} else {
			break
		}
	}
	for {
		err := db.Ping()
		if err != nil {
			lg.Printf("Ping failed: '%s'", err)
			lg.Printf("Retrying after 5 seconds...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	lg.Printf("Connected to postgres successfully")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		lg.Printf("Creating postgres driver failed: '%s'", err)
		os.Exit(1)
	}
	lg.Printf("Creating postgres driver succeeded")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver)
	if err != nil {
		lg.Printf("Creating migration failed: '%s'", err)
		os.Exit(1)
	}
	if m == nil {
		lg.Printf("Creating migration failed: nil pointer")
		os.Exit(1)
	}
	defer m.Close()

	lg.Printf("Creating migration succeeded")

	// Drop all tables.
	if *fresh {
		err = m.Down()
		if err != nil {
			lg.Printf("Migration: Down() failed: '%s'", err)
			os.Exit(1)
		}
		lg.Printf("Migration: Down() succeeded!")
	}
	// User-defined force, doesn't matter if dirty or not
	if forceStep >= 0 {
		err = m.Force(forceStep)
		if err != nil {
			lg.Printf("Migration: Force(%d) failed: '%s'", forceStep, err)
			os.Exit(1)
		}
		lg.Printf("Migration: Force(%d) succeeded!", forceStep)
	}
	noVersion := false
	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		lg.Printf("Migration: Version: No version (%d) yet", version)
		noVersion = true
	} else if err != nil {
		lg.Printf("Migration: Version() failed unexpectedly: '%s'", err)
		os.Exit(1)
	} else {
		lg.Printf("Migration: At step version %d, dirty: %t", version, dirty)
	}
	if dirty && forceStep < 0 {
		// Force current version to remove dirty flag.  We'd prefer to avoid
		// this situation in the first place.
		err = m.Force(int(version))
		if err != nil {
			lg.Printf("Migration: (Dirty) Force(%d) failed: '%s'", version, err)
			os.Exit(1)
		}
		lg.Printf("Migration: (Dirty) Force(%d) succeeded!", version)
	}
	// Initial install - migrate all the way up.
	if noVersion {
		lg.Printf("Migration: Initial install, call Up()...")
		err = m.Up()
		if err == migrate.ErrNoChange {
			lg.Printf("Migration: Up(): WARNING: No changes required?")
		} else if err != nil {
			lg.Printf("Migration: Up() failed: '%s'", err)
			os.Exit(1)
		} else {
			lg.Printf("Migration: Up() succeeded!")
		}
	} else if version != migrateStep {
		if version < migrateStep {
			lg.Printf("Migration: DB at step %d/%d. Updating...", version, migrateStep)
		} else {
			lg.Printf("Migration: DB at step %d/%d. Downgrading...", version, migrateStep)
		}
		err = m.Migrate(migrateStep)
		if err == migrate.ErrNoChange {
			lg.Printf("Migration: Migrate(%d): No changes required?", migrateStep)
		} else if err != nil {
			lg.Printf("Migration: Migrate(%d) failed: '%s'", migrateStep, err)
			os.Exit(1)
		} else {
			lg.Printf("Migration: Migrate(%d) succeeded!", migrateStep)
		}
	} else {
		lg.Printf("Migration: Already at expected step.  Nothing to do.")
		os.Exit(0)
	}
	version2, dirty2, err := m.Version()
	if err == migrate.ErrNilVersion {
		lg.Printf("Migration: NO VERSION AFTER MIGRATE (%d) yet", version2)
	} else if err != nil {
		lg.Printf("Migration: Version() failed unexpectedly: '%s'", err)
		os.Exit(1)
	} else {
		lg.Printf("Migration: At step version %d, dirty: %t", version2, dirty2)
	}
}
