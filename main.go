package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/subgraph/usblockout/config"
	mlog "github.com/subgraph/usblockout/logging"

	"github.com/godbus/dbus"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger(config.AppName)

const (
	dbusObjectScreensaverActiveChange = "type='signal',path='/org/gnome/ScreenSaver',interface='org.gnome.ScreenSaver',member='ActiveChanged'"
)

var dbusObjectSetLocked = config.BusName + ".SetLocked"
var dbusObjectIsRunning = config.BusName + ".IsRunning"

type usbLockout struct {
	dconnSys  *dbus.Conn
	dbusSys   dbus.BusObject
	dconnSess *dbus.Conn
	logBackend logging.LeveledBackend
}

func (ul *usbLockout) lock() error {
	if res := ul.dbusSys.Call(dbusObjectSetLocked, 0, true); res.Err != nil {
		return res.Err
	}
	return nil
}

func (ul *usbLockout) unlock() error {
	if res := ul.dbusSys.Call(dbusObjectSetLocked, 0, false); res.Err != nil {
		return res.Err
	}
	return nil
}

func (ul *usbLockout) processSignals(c <-chan os.Signal) {
	for {
		sig := <-c
		log.Debugf("Recieved signal (%v)\n", sig)
		if err := ul.lock(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func (ul *usbLockout) runDaemon() {
	for {
		c := make(chan *dbus.Signal, 1)
		ul.dconnSess.Signal(c)
		for v := range c {
			state := v.Body[0].(bool)
			switch state {
			case true:
				log.Debug("Screen locked")
				if err := ul.lock(); err != nil {
					log.Fatal(err)
					os.Exit(1)
				}
			case false:
				log.Debug("Screen unlocked")
				if err := ul.unlock(); err != nil {
					log.Fatal(err)
					os.Exit(1)
				}
				break
			default:
				log.Errorf("Unknown signal received: %+v\n", v.Body[0])
			}
			break
		}
		ul.dconnSess.RemoveSignal(c)
	}
}

var flagdebug, enable, disable bool

func init() {
	flag.BoolVar(&enable, "enable", false, "manually enable the usb deny feature")
	flag.BoolVar(&disable, "disable", false, "manually disable the usb deny feature")
	flag.BoolVar(&flagdebug, "debug", false, "enable debug logging")
	flag.Parse()
}

func main() {
	var err error
	ul := &usbLockout{}
	ul.logBackend = mlog.SetupLoggerBackend(logging.INFO, config.AppName)
	log.SetBackend(ul.logBackend)
	if flagdebug {
		ul.logBackend.SetLevel(logging.DEBUG, config.AppName)
		log.Debug("Debug logging enabled")
	}

	ul.dconnSys, err = dbus.SystemBus()
	if err != nil {
		log.Errorf("Failed to connect to system bus: %+v\n", err)
		os.Exit(1)
	}

	ul.dconnSess, err = dbus.SessionBus()
	if err != nil {
		log.Errorf("Failed to connect to session bus: %+v\n", err)
		os.Exit(1)
	}

	ul.dconnSess.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, dbusObjectScreensaverActiveChange)

	ul.dbusSys = ul.dconnSys.Object(config.BusName, config.ObjectPath)

	if res := ul.dbusSys.Call(dbusObjectIsRunning, 0); res.Err != nil || res.Body[0] == false {
		log.Error("USB Lockout daemon is not running or unavailable")
		os.Exit(1)
	}

	switch {
	case enable == true:
		fmt.Println("Enabling USB Deny")
		ul.lock()
		os.Exit(0)
		break
	case disable == true:
		fmt.Println("Disabling USB Deny")
		ul.unlock()
		os.Exit(0)
		break
	default:
		log.Notice("USB Lockout client enabled")
		if res := ul.dbusSys.Call(dbusObjectSetLocked, 0, false); res.Err != nil {
			log.Fatal(res.Err)
			os.Exit(1)
		}

		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
		go ul.processSignals(sigs)

		ul.runDaemon()
	}
}
