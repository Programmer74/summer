package main

import (
	log "github.com/jeanphorn/log4go"
	"github.com/programmer74/summer/summer"
	"github.com/programmer74/summer/testPackage"
)

func main() {
	log.LoadConfiguration("./log4go.json")
	defer log.Close()

	summer.ParseProperties("./example.properties")

	summer.RegisterBean("1", testPackage.C{Z: 3})
	summer.RegisterBean("2", testPackage.A{X: 2})
	summer.RegisterBean("3", testPackage.B{Y: 1})
	summer.RegisterBeanWithTypeAlias("4", testPackage.Impl2{}, "*testPackage.Ifce")
	summer.RegisterBean("customBeanName", testPackage.A{X: 5})

	summer.PerformDependencyInjection()

	a := summer.GetBean("2").(*testPackage.A)
	b := summer.GetBean("3").(*testPackage.B)
	c := summer.GetBean("1").(*testPackage.C)

	_ = a
	_ = b

	log.Info("Answer is %d", c.DoC())
	log.Info("Property is %d", c.C)
	c.DoInterfaceSpecificStuff()
	log.Info("BeanByCustomName stuff is %d", c.DoCustomBeanNameInjectedStuff())
}
