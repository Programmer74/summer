package testPackage

type C struct {
	Z int
	A *interface{} `summer:"*testPackage.A"`
	B *interface{} `summer:"*testPackage.B"`
	C int          `summer.property:"testPropertyName|123"`
	D *interface{} `summer:"*testPackage.Ifce"`
	E *interface{} `summer:"customBeanName"`
}

func (c *C) getA() *A {
	a := c.A
	a2 := (*a).(*A)
	return a2
}

func (c *C) getB() *B {
	b := c.B
	b2 := (*b).(*B)
	return b2
}

func (c *C) getD() Ifce {
	d := *c.D
	d2 := (d).(Ifce)
	return d2
}

func (c *C) getE() *A {
	a := c.E
	a2 := (*a).(*A)
	return a2
}

func (c *C) DoC() int {
	return 100*c.Z + c.getB().DoB() + c.getA().DoA()
}

func (c *C) DoInterfaceSpecificStuff() {
	c.getD().DoIfceStuff()
}

func (c *C) DoCustomBeanNameInjectedStuff() int {
	return 3000 + c.getE().DoA()
}