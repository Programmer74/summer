package testPackage

import log "github.com/jeanphorn/log4go"

type C struct {
	Z int
	A *interface{} `summer:"testPackage.A"`
	b *B
	f *Ifce
}

func (c *C) DoC() {
	log.Warn("c did it")
}