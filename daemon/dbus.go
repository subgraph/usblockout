package daemon

import (
	"errors"
	"fmt"
	"strings"

	"github.com/subgraph/usblockout/config"

	"github.com/godbus/dbus"
)

type dbusServer struct {
	ul   *usbLockoutd
	conn *dbus.Conn
}

func newDbusServer() (*dbusServer, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	reply, err := conn.RequestName(config.BusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.New("Bus name is already owned")
	}
	ds := &dbusServer{}

	if err := conn.Export(ds, config.ObjectPath, config.InterfaceName); err != nil {
		return nil, err
	}

	ps := strings.Split(config.ObjectPath, "/")
	path := "/"
	for _, p := range ps {
		if len(path) > 1 {
			path += "/"
		}
		path += p

		if err := conn.Export(ds, dbus.ObjectPath(path), "org.freedesktop.DBus.Introspectable"); err != nil {
			return nil, err
		}
	}
	ds.conn = conn
	return ds, nil
}

func (ds *dbusServer) Introspect(msg dbus.Message) (string, *dbus.Error) {
	path := string(msg.Headers[dbus.FieldPath].Value().(dbus.ObjectPath))
	if path == config.ObjectPath {
		return config.IntrospectXML, nil
	}
	parts := strings.Split(config.ObjectPath, "/")
	current := "/"
	for i := 0; i < len(parts)-1; i++ {
		if len(current) > 1 {
			current += "/"
		}
		current += parts[i]
		if path == current {
			next := parts[i+1]
			return fmt.Sprintf("<node><node name=\"%s\"/></node>", next), nil
		}
	}
	return "", nil
}

func (ds *dbusServer) SetLocked(flag bool) *dbus.Error {
	log.Debugf("SetLocked(%v) called", flag)
	ds.ul.setLocked(flag)
	return nil
}

func (ds *dbusServer) IsRunning() (bool, *dbus.Error) {
	log.Debug("IsRunning() called")
	return true, nil
}
