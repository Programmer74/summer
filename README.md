# summer

A Spring-like IoC/DI framework (or actually a pet project) for Go

## Warning
Since the Go's reflection is not as powerful as one in Java, I had to establish the following rules:
- Autowiring only ```*interface{}``` (like using Java's Object everywhere); therefore you have to cast everything or create a casting getter (which is what I used);
- Autowiring via field injection only (though I may change it in future, since it seems I can call methods by name via Go reflection);
- When you need to autowire an interface, you have to specify that manually.
## Usage

See the [example](testExample.go).

## Features

- Injection by name:
```
type C struct {
	E *interface{} `summer:"customBeanName"`
}

summer.RegisterBean("customBeanName", testPackage.A2{})
```

- Injection by interface:
```
type C struct {
	D *interface{} `summer:"*testPackage.Ifce"`
}

type Ifce interface {
	DoIfceStuff()
}

type Impl2 struct{}

func (i *Impl2) DoIfceStuff() {
	log.Warn("Impl2")
}

summer.RegisterBeanWithTypeAlias("beanName", testPackage.Impl2{}, "*testPackage.Ifce")
```

- Injection by type:
```
type C struct {
    A *interface{} `summer:"*testPackage.A"`
}

summer.RegisterBean("beanName", testPackage.A{})
```

- Injecting properties:
```
type C struct {
    C int `summer.property:"testPropertyName|123"`
}

summer.ParseProperties("./example.properties")
```

- Dependency graph printer
