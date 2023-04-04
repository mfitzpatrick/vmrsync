//go:build windows && service

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	SVC_NAME  = "VMRSync"
	SVC_DNAME = "VMR TripWatch Synchronisation"
	SVC_DESC  = "TripWatch and VMR Southport DB Synchronisation Service"
)

// Main run loop which is suitable for use by a Windows Service.
func runLoop() {
	if inService, err := svc.IsWindowsService(); err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	} else if inService {
		// We are currently running as a Windows Service. Execute the main service
		// runtime loop.
		runService(SVC_NAME, false)

		// Return early if we exit from the main runtime loop, as we don't want to reinstall
		// ourselves.
		return
	}

	// We are currently not running as a service. First check if a pre-existing service exists
	// for us. If it does stop and delete it.
	if err := removeService(SVC_NAME); err != nil && !errors.Is(err, svcNotInstalledError) {
		// Is an error, but isn't a "not-installed error". Must be fatal.
		log.Fatalf("Failed to remove service %s: %v", SVC_NAME, err)
	}

	// Now install ourselves as a service, and start us running.
	if err := installService(SVC_NAME); err != nil {
		log.Fatalf("Failed to install service %s: %v", SVC_NAME, err)
	} else if err := startService(SVC_NAME); err != nil {
		log.Fatalf("Failed to start service %s: %v", SVC_NAME, err)
	}
}

var svcNotInstalledError error = errors.New("windows service not installed")
var elog debug.Log

type winsvc struct {
	db *sql.DB
}

func (s *winsvc) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	// Open the config file here because we need this information in order to properly configure the ticks.
	// But we don't want to connect to the DB at this time, because it makes our app unresponsive.
	if err := parseConfig(configFilePath); err != nil {
		elog.Error(1, fmt.Sprintf("VMRSync failed to open config: %v", err))
		return
	}
	fasttick := time.Tick(tripwatchPollFrequency)
	slowtick := time.Tick(4 * tripwatchPollFrequency)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, fmt.Sprintf("VMRSync %s execution started. Poll frequency: %s",
		Version, tripwatchPollFrequency))

loop:
	for {
		select {
		case <-tick:
			if s.db == nil {
				if db, closefunc, err := setup(); err != nil {
					elog.Error(1, fmt.Sprintf("Cannot connect to DB: %v", err))
				} else {
					defer closefunc()
					s.db = db
				}
			}
			if errlist := run(s.db); len(errlist) > 0 {
				for _, err := range errlist {
					if errors.Is(err, matchFieldIsZero) {
						var runerr runError
						if ok := errors.As(err, &runerr); ok {
							elog.Error(1, fmt.Sprintf("Couldn't match field for %s",
								runerr.String()))
						} else {
							elog.Error(1, "Missing match field (and runError object)")
						}
					} else {
						elog.Error(1, fmt.Sprintf("Run loop failure: %+v", err))
					}
				}
			}
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				elog.Info(1, "shutdown has been requested")
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	if err = run(name, &winsvc{db: nil}); err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", errors.Wrapf(err, "exe path finding abspath")
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = errors.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = errors.Errorf("%s is directory", p)
		}
	}
	return "", errors.Wrapf(err, "exe path")
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return errors.Wrapf(err, "start service connecting")
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return errors.Wrapf(err, "could not access service")
	}
	defer s.Close()
	err = s.Start()
	if err != nil {
		return errors.Wrapf(err, "could not start service")
	}
	return nil
}

func installService(name string) error {
	exepath, err := exePath()
	if err != nil {
		return errors.Wrapf(err, "installing service finding exe path")
	}
	m, err := mgr.Connect()
	if err != nil {
		return errors.Wrapf(err, "installing service connecting")
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return errors.Errorf("service %s already exists", name)
	}
	s, err = m.CreateService(name, exepath, mgr.Config{
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
		DisplayName:      SVC_DNAME,
		Description:      SVC_DESC,
	}, "-config-file", filepath.Join(filepath.Dir(exepath), ".config.yml"))
	if err != nil {
		return errors.Wrapf(err, "installing service couldn't create it")
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return errors.Wrapf(err, "SetupEventLogSource() failed")
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return errors.Wrapf(err, "removing service connecting")
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return errors.Wrapf(svcNotInstalledError, "service %s is not installed (internal err %v)",
			name, err)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return errors.Wrapf(err, "removing service performing delete operation")
	}
	err = eventlog.Remove(name)
	if err != nil {
		return errors.Wrapf(err, "RemoveEventLogSource() failed")
	}
	return nil
}
