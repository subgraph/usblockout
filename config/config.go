package config

import (
	"github.com/godbus/dbus/introspect"
)

const IntrospectXml = `
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
	BusName       = "com.subgraph.USBLockout"
	ObjectPath    = "/com/subgraph/USBLockout"
	InterfaceName = "com.subgraph.USBLockout"
	AppName       = "usblockout"
)
