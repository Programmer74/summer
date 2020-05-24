package main

import (
	log "github.com/jeanphorn/log4go"
	"github.com/programmer74/summer/summer"
	"github.com/programmer74/summer/testPackage"
)

func main() {
	log.LoadConfiguration("./log4go.json")
	defer log.Close()

	summer.RegisterBean("1", testPackage.C{Z: 3})
	summer.RegisterBean("2", testPackage.A{X: 1})
	summer.RegisterBean("3", testPackage.B{Y: 2})
	summer.RegisterBean("4", testPackage.Impl1{})
	summer.PerformDependencyInjection()

	c := summer.GetBean("1").(*testPackage.C)
	c.DoC()
}
