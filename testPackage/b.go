package testPackage

type B struct {
	Y int
	A *interface{} `summer:"*testPackage.A"`
}

func (b *B) getA() *A {
	a := b.A
	a2 := (*a).(*A)
	return a2
}

func (b *B) DoB() int {
	return 10 * b.Y + b.getA().DoA()
}