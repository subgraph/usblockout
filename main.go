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
	DbusObjectScreensaverActiveChange = "type='signal',path='/org/gnome/ScreenSaver',interface='org.gnome.ScreenSaver',member='ActiveChanged'"
)

var DbusObjectSetLocked = config.BusName + ".SetLocked"
var DbusObjectIsRunning = config.BusName + ".IsRunning"

type USBLockout struct {
	dconn_sys  *dbus.Conn
	dbus_sys   dbus.BusObject
	dconn_sess *dbus.Conn
	logBackend logging.LeveledBackend
}

func (ul *USBLockout) lock() error {
	if res := ul.dbus_sys.Call(DbusObjectSetLocked, 0, true); res.Err != nil {
		return res.Err
	}
	return nil
}

func (ul *USBLockout) unlock() error {
	if res := ul.dbus_sys.Call(DbusObjectSetLocked, 0, false); res.Err != nil {
		return res.Err
	}
	return nil
}

func (ul *USBLockout) processSignals(c <-chan os.Signal) {
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

func (ul *USBLockout) runDaemon() {
	for {
		c := make(chan *dbus.Signal, 1)
		ul.dconn_sess.Signal(c)
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
		ul.dconn_sess.RemoveSignal(c)
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
	ul := &USBLockout{}
	ul.logBackend = mlog.SetupLoggerBackend(logging.INFO, config.AppName)
	log.SetBackend(ul.logBackend)
	if flagdebug {
		ul.logBackend.SetLevel(logging.DEBUG, config.AppName)
		log.Debug("Debug logging enabled")
	}

	ul.dconn_sys, err = dbus.SystemBus()
	if err != nil {
		log.Errorf("Failed to connect to system bus: %+v\n", err)
		os.Exit(1)
	}

	ul.dconn_sess, err = dbus.SessionBus()
	if err != nil {
		log.Errorf("Failed to connect to session bus: %+v\n", err)
		os.Exit(1)
	}

	ul.dconn_sess.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, DbusObjectScreensaverActiveChange)

	ul.dbus_sys = ul.dconn_sys.Object(config.BusName, config.ObjectPath)

	if res := ul.dbus_sys.Call(DbusObjectIsRunning, 0); res.Err != nil || res.Body[0] == false {
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
		if res := ul.dbus_sys.Call(DbusObjectSetLocked, 0, false); res.Err != nil {
			log.Fatal(res.Err)
			os.Exit(1)
		}

		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
		go ul.processSignals(sigs)

		ul.runDaemon()
	}
}
