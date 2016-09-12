package daemon

import (
	"flag"
	"os"
	"runtime"
	"sync"

	mlog "github.com/subgraph/usblockout/logging"
	"github.com/subgraph/usblockout/daemon/sysctl"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("usblockout")

type USBLockoutd struct {
	dbus       *dbusServer
	locked     bool
	lock       sync.Mutex
	logBackend logging.LeveledBackend
}

const (
	KERN_GRSEC_DENY_NEW_USB = "kernel.grsecurity.deny_new_usb"
)

func (ul *USBLockoutd) setLocked(flag bool) error {
	ul.lock.Lock()
	defer ul.lock.Unlock()
	ul.locked = flag
	ival := "0"
	if ul.locked {
		ival = "1"
	}
	if err := sysctl.Set(KERN_GRSEC_DENY_NEW_USB, ival); err != nil {
		log.Errorf("Error setting grsec deny new usb: %+v", err)
		return err
	}

	if str, err := sysctl.Get(KERN_GRSEC_DENY_NEW_USB); err != nil {
		log.Warningf("%s: %s > %+v", "KERN_GRSEC_DENY_NEW_USB", str, err)
	} else {
		log.Noticef("%s: %s", "KERN_GRSEC_DENY_NEW_USB", str)
	}

	return nil
}

var flagdebug bool
func init() {
	flag.BoolVar(&flagdebug, "debug", false, "enable debug logging")
	flag.Parse()
}

func Main() {
	logBackend := mlog.SetupLoggerBackend(logging.INFO, "usblockout")
	log.SetBackend(logBackend)
	if flagdebug {
		logBackend.SetLevel(logging.DEBUG, "usblockout")
		log.Debug("Debug logging enabled")
	}

	if os.Geteuid() != 0 || runtime.GOOS != "linux" {
		log.Error("Must be run as root")
		os.Exit(1)
	}

	ds, err := newDbusServer()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	ul := &USBLockoutd{
		dbus:   ds,
		locked: false,
	}

	if err := ul.setLocked(true); err != nil {
		log.Fatalf("unable to write to sysctl: %+v", err)
		os.Exit(1)
	}

	ds.ul = ul

	log.Notice("USB Lockout daemon enabled")
	select {}
}

// TODO: Handle exit signal (enable deny usb?)
