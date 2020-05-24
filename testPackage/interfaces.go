package testPackage

type Ifce interface {
	DoIfceStuff()
}

type Impl1 struct {}
func (i *Impl1) DoIfceStuff() {
	println("Impl1")
}

type Impl2 struct {}
func (i *Impl2) DoIfceStuff() {
	println("Impl2")
}