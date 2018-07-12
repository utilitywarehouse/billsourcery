// +build windows

package main

import "os/user"

func defaultCacheName() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir + "/billsourcery_timestats_cache"
}

func defaultOutputName() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir + "/billtimestats.png"
}
