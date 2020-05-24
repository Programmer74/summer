package testPackage

type B struct {
	Y int
	a *A
}

func (b *B) doB() {
	println("b did it")
}