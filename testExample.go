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

	summer.RegisterBean("BeanC", testPackage.C{Z: 3})
	summer.RegisterBean("BeanA", testPackage.A{X: 2})
	summer.RegisterBean("BeanB", testPackage.B{Y: 1})
	summer.RegisterBeanWithTypeAlias("Impl2:Ifce", testPackage.Impl2{}, "*testPackage.Ifce")
	summer.RegisterBean("customBeanName", testPackage.A{X: 5})

	summer.PerformDependencyInjection()

	a := summer.GetBean("BeanA").(*testPackage.A)
	b := summer.GetBean("BeanB").(*testPackage.B)
	c := summer.GetBean("BeanC").(*testPackage.C)

	_ = a
	_ = b

	log.Info("Answer is %d", c.DoC())
	log.Info("Property is %d", c.C)
	c.DoInterfaceSpecificStuff()
	log.Info("BeanByCustomName stuff is %d", c.DoCustomBeanNameInjectedStuff())

	summer.PrintDependencyGraphVertex()
}
