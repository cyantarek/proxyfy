// +build !windows
package privileges

import (
	"os/user"
	"strconv"
	"syscall"

	"proxyfy/pkg/logger"
)

// DropPrivilege changes the uid/gid
func DropPrivilege(uids, gids string) {
	if me := syscall.Getuid(); me != 0 {
		logger.Log.Warningln("Not running as 'root'; can't change uid/gid")
		return
	}

	if len(gids) > 0 {
		gi, err := user.LookupGroup(gids)
		if err != nil {
			gi, err = user.LookupGroupId(gids)
			if err != nil {
				logger.Log.Fatalf("can't find group '%s' to drop privilege: %s", gids, err)
			}
		}

		gid, err := strconv.Atoi(gi.Gid)
		if err != nil {
			logger.Log.Fatalf("can't parse integer gid %s: %s", gi.Gid, err)
		}

		if err = syscall.Setgid(gid); err != nil {
			logger.Log.Fatalf("can't change Gid to %d: %s", gid, err)
		}
	}

	if len(uids) > 0 {
		ui, err := user.Lookup(uids)
		if err != nil {
			ui, err = user.LookupId(uids)
			if err != nil {
				logger.Log.Fatalf("can't find user '%s' to drop privilege: %s", uids, err)
			}
		}
		uid, err := strconv.Atoi(ui.Uid)
		if err != nil {
			logger.Log.Fatalf("can't parse integer uid %s: %s", ui.Uid, err)
		}

		if err = syscall.Setuid(uid); err != nil {
			logger.Log.Fatalf("can't change Uid to %d: %s", uid, err)
		}
	}
}
