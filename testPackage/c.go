package testPackage

type C struct {
	Z int
	A *interface{} `summer:"*testPackage.A"`
	B *interface{} `summer:"*testPackage.B"`
	f *Ifce
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

func (c *C) DoC() int {
	return 100 * c.Z + c.getB().DoB() + c.getA().DoA();
}