package profiler

var (
	GenericPathsSet = map[string]bool{
		"/": true, "/bin": true, "/boot": true, "/boot/efi": true, "/dev": true, "/dev/pts": true, "/dev/shm": true, "/etc": true,
		"/etc/network": true, "/etc/opt": true, "/etc/ssl": true, "/home": true, "/lib": true, "/lib32": true, "/lib64": true,
		"/lib/firmware": true, "/lib/x86_64-linux-gnu": true, "/media": true, "/mnt": true, "/opt": true, "/proc": true,
		"/root": true, "/run": true, "/run/lock": true, "/run/shm": true, "/sbin": true, "/srv": true, "/sys": true, "/tmp": true,
		"/usr": true, "/usr/bin": true, "/usr/games": true, "/usr/include": true, "/usr/lib": true, "/usr/lib64": true,
		"/usr/libexec": true, "/usr/lib/locale": true, "/usr/local": true, "/usr/local/bin": true,
		"/usr/local/games": true, "/usr/local/lib": true, "/usr/local/lib64": true, "/usr/local/sbin": true,
		"/usr/sbin": true, "/usr/share": true, "/usr/share/doc": true, "/usr/share/fonts": true,
		"/usr/share/icons": true, "/usr/share/locale": true, "/usr/share/man": true, "/usr/share/themes": true,
		"/var": true, "/var/backups": true, "/var/cache": true, "/var/lib": true, "/var/lib/apt": true,
		"/var/lib/dhcp": true, "/var/lib/dpkg": true, "/var/lib/snapd": true, "/var/lib/systemd": true,
		"/var/lock": true, "/var/log": true, "/var/mail": true, "/var/opt": true, "/var/run": true, "/var/spool": true,
		"/var/tmp": true, "/var/www": true, "/usr/local/bin/bash": true, "/usr/local/sbin/bash": true,
		"/usr/sbin/bash": true, "/usr/bin/bash": true, "/usr/lib/x86_64-linux-gnu": true,
	}

	ExcludePrefixesSet = map[string]bool{
		"/dev/": true, "/proc/": true, "/sys/": true, "/run/": true, "/tmp/": true, "/usr/lib/locale/": true, "/usr/share/locale/": true,
	}
)
