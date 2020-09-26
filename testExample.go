package main

import (
	log "github.com/jeanphorn/log4go"
	"github.com/nikita-tomilov/summer/summer"
	"github.com/nikita-tomilov/summer/testPackage"
)

func main() {
	log.LoadConfiguration("./log4go.json")
	defer log.Close()

	summer.ParseProperties("./example.properties")

	value, _ := summer.GetPropertyValue("anotherProperty")
	log.Info("Property value: %s", value)

	summer.RegisterBean("BeanC", testPackage.C{Z: 3})
	summer.RegisterBean("BeanA", testPackage.A{X: 2})
	summer.RegisterBean("BeanB", testPackage.B{Y: 1})
	summer.RegisterBeanWithTypeAlias("Impl2:Ifce", testPackage.Impl2{}, "*testPackage.Ifce")
	summer.RegisterBean("customBeanName", testPackage.A2{X: 5})

	summer.PerformDependencyInjection()

	a := (*summer.GetBean("BeanA")).(*testPackage.A)
	b := (*summer.GetBean("BeanB")).(*testPackage.B)
	c := (*summer.GetBean("BeanC")).(*testPackage.C)

	_ = a
	_ = b

	log.Info("Answer is %d", c.DoC())
	log.Info("Property is %d", c.C)
	c.DoInterfaceSpecificStuff()
	log.Info("BeanByCustomName stuff is %d", c.DoCustomBeanNameInjectedStuff())

	summer.PrintDependencyGraphVertex()
}
