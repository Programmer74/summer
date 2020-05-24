package testPackage

type A struct {
	X int
}

func (a *A) DoA() int {
	return a.X
}