# Session USB Lockout

This program provides a way to toggle Grsecurity Deny New USB feature with the state of a user session.
That is, it will automatically enable the feature when the screen is locked or the session exits, and vice versa.

## Caveats

**Beware! If you use some sort of USB device (ex: a yubikey) for pam logins, login will be entierly broken!**
One workaround for this is to plug in the usb device at boot (before the deamon launches), or before switching to a different tty.

This, of course only works if Grsecurity sysctl are enabled and is not locked.

## Building & Packaging

Provided in this repository is a debian branch which is used to build a deb package from git tags:

	git checkout -b debian https://github.com/subgraph/usblockout.git
	cd usblockout
	gbp buildpackage -us -uc
	dpkg -i /tmp/subgraph-usblockout_#VERSION#.deb

You will need to either log out and log back in, or launch `usblockout` (for example via alt-f2) after the install.

To run without without the xdg autostart and systemd service (when you are debugging or for development) you will want to run the daemon in one terminal with `sudo ./usblockoutd --debug` and the client in another with `./usblockout --debug`.
