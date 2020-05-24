package testPackage

import log "github.com/jeanphorn/log4go"

type Ifce interface {
	DoIfceStuff()
}

type Impl1 struct{}

func (i *Impl1) DoIfceStuff() {
	log.Warn("Impl1")
}

type Impl2 struct{}

func (i *Impl2) DoIfceStuff() {
	log.Warn("Impl2")
}
