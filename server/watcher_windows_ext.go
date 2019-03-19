// +build windows

package main

import (
	"github.com/rjeczalik/notify"
)

func init() {
	WatcherNotifies = []notify.Event{notify.All, notify.FileNotifyChangeLastWrite}
}
