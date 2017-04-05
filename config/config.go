package config

import (
	"github.com/godbus/dbus/introspect"
)

// IntrospectXML contains the dbus node definition
const IntrospectXML = `
<node>
  <interface name="com.subgraph.USBLockout">
    <method name="SetLocked">
      <arg name="locked" direction="in" type="b" />
    </method>
    <method name="IsRunning">
      <arg name="running" direction="out" type="b" />
    </method>
  </interface>` +
	introspect.IntrospectDataString +
	`</node>`

const (
	// BusName contains the dbus node name
	BusName       = "com.subgraph.USBLockout"
	// ObjectPath contains the dbus object path
	ObjectPath    = "/com/subgraph/USBLockout"
	// InterfaceName contains the dbus interface name
	InterfaceName = "com.subgraph.USBLockout"
	// AppName is the application name
	AppName       = "usblockout"
)
