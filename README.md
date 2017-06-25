# Session USB Lockout

You're in a place that is relatively safe, say your local hackerspace.
You need to leave your computer out of sight for a small amount of time,
the space is safe enough that you are not worried about it being stolen,
but not enough that someone couldn't attempt a quick drive-by USB attack.

This program provides a way to toggle [Grsecurity](https://grsecurity.net/) [Deny New USB feature](https://en.wikibooks.org/wiki/Grsecurity/Appendix/Grsecurity_and_PaX_Configuration_Options#Deny_new_USB_connections_after_toggle) with the state of a user session.
That is, it will automatically enable the feature when the screen is locked or the session exits, and vice versa.

It consists of a privileged daemon that exposes itself on the dbus system bus;
and a client daemon which runs in the user session via xdg-autostart, and relays the session screen-lock events on the system bus.

The client utility also allows the user to enable or disable the feature manually by calling:

	usblockout --[enable|disable]

## Caveats

**Beware! If you use some sort of USB device (ex: a YubiKey) for PAM logins, login will be entirely broken!**
One workaround for this is to plug in the USB device at boot (before the daemon launches), or before switching to a different tty.
Devices like smartcards, which have readers that are always plugged in, should work as expected.

This, of course only works if Grsecurity sysctl is enabled and not locked.

## Building & Packaging

Provided in this repository is a debian branch which is used to build a deb package from git tags:

	git checkout -b debian https://github.com/subgraph/usblockout.git
	cd usblockout
	gbp buildpackage -us -uc
	dpkg -i /tmp/subgraph-usblockout_#VERSION#.deb

You will need to either log out and log back in, or launch `usblockout` (for example via alt-F2) after the install.

To run without without the xdg autostart and systemd service (when you are debugging or for development) you will want to run the daemon in one terminal with `sudo ./usblockoutd --debug` and the client in another with `./usblockout --debug`.
