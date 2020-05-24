package summer

import (
	"errors"
	"fmt"
	log "github.com/jeanphorn/log4go"
	"reflect"
	"strings"
)

var uninitializedBeans = make(map[string]interface{})

var initializedBeansList = make([]interface{}, 0)
var initializedBeanNamesList = make([]string, 0)

func RegisterBean(beanName string, bean interface{}) {
	uninitializedBeans[beanName] = bean
}

func GetBean(beanName string) interface{} {
	for i := 0; i < len(initializedBeansList); i++ {
		if initializedBeanNamesList[i] == beanName {
			return initializedBeansList[i]
		}
	}
	panic("no processed bean found")
}

func PerformDependencyInjection() {
	oneBeanInitialized := true
	for oneBeanInitialized {
		oneBeanInitialized = false

		for beanName, bean := range uninitializedBeans {
			//println("===bean", beanName, ":")

			//beanType := reflect.TypeOf(bean)
			//examiner(beanType, 0)

			annotatedDependenciesCount := getAnnotatedDependenciesCount(bean)
			if annotatedDependenciesCount == 0 {
				log.Debug("Bean %s has no summer dependencies", getString(beanName, bean))
				oneBeanInitialized = advanceBeanAsInitialized(beanName, bean)
			} else {
				log.Debug("Bean %s has summer dependencies", getString(beanName, bean))
				unprocessedDependenciesLeft, processedBean := tryFillDependencies(beanName)
				if unprocessedDependenciesLeft > 0 {
					log.Debug("Bean %s has %d unprocessed summer dependencies left", getString(beanName, bean), unprocessedDependenciesLeft)
				} else {
					log.Debug("Bean %s has no unprocessed summer dependencies left", getString(beanName, bean))
					oneBeanInitialized = advanceBeanAsInitialized(beanName, processedBean)
				}
			}
		}
	}

	if len(uninitializedBeans) > 0 {
		for beanName, bean := range uninitializedBeans {
			log.Critical("BEAN %s WAS NOT INITIALIZED", getString(beanName, bean))
		}
		panic("UninializedBeans size != 0")
	}
}

func getType(x interface{}) string {
	t := reflect.TypeOf(x)
	return t.Name() + " " + t.Kind().String()
}

func getString(beanName string, bean interface{}) string {
	return "'" + beanName + "' (" + getType(bean) + ")"
}

func advanceBeanAsInitialized(beanName string, bean interface{}) bool {
	initializedBeansList = append(initializedBeansList, bean)
	initializedBeanNamesList = append(initializedBeanNamesList, beanName)

	delete(uninitializedBeans, beanName)
	return true
}

func getAnnotatedDependenciesCount(x interface{}) int {
	t := reflect.TypeOf(x)
	annotatedDependenciesCount := 0
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		return 0
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Tag != "" {
				_, ok := f.Tag.Lookup("summer")
				if ok {
					annotatedDependenciesCount++
				}
			}
		}
	}
	return annotatedDependenciesCount
}

//func tryFillDependencies(x *interface{}, sourceType reflect.Type) int {
func tryFillDependencies(beanName string) (int, interface{}) {
	summerTotalDependenciesCount := 0
	summerInjectedDependenciesCount := 0

	x := uninitializedBeans[beanName]
	sourceType := reflect.TypeOf(x)

	xc := reflect.New(sourceType)
	xCopy := xc.Interface().(interface{})

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		if sourceField.Tag != "" {
			requiredTypeAsString, ok := sourceField.Tag.Lookup("summer")
			if ok {
				log.Debug("Injectable 'summer' field %s", sourceField)
				summerTotalDependenciesCount++

				targetField := reflect.ValueOf(xCopy).Elem().Field(i)
				if targetField.IsValid() {
					if targetField.Kind() != reflect.Ptr {
						panic("field annotated by 'summer' should be pointers")
					}
					if targetField.IsNil() {
						log.Debug("Field %s is not set", targetField)
						applicableBean, err := getProcessedBeanByType(requiredTypeAsString)
						if err == nil {
							log.Debug("Field %s value found : %s", targetField, applicableBean)
							if !targetField.CanSet() {
								log.Critical("CANNOT SET FIELD %s", targetField)
							}

							targetField.Set(reflect.ValueOf(applicableBean).Convert(targetField.Type()))
							summerInjectedDependenciesCount++
						} else {
							log.Error("Field %s value not found by now", targetField)
						}
					} else {
						log.Debug("Field %s already set", targetField)
						summerInjectedDependenciesCount++
					}
				} else {
					log.Error("For some reason, %s is not valid", targetField)
				}
			}
		}
	}

	unprocessedDependenciesLeft := summerTotalDependenciesCount - summerInjectedDependenciesCount
	return unprocessedDependenciesLeft, xCopy
}

//func getProcessedBeanByType(requiredType reflect.Type) (*interface{}, error) {
func getProcessedBeanByType(requiredTypeAsString string) (*interface{}, error) {
	log.Debug("  Asked to find '%s'", requiredTypeAsString)

	compatibleBeansIndexes := make([]int, 0)

	for i := 0; i < len(initializedBeansList); i++ {
		bean := initializedBeansList[i]
		beanName := initializedBeanNamesList[i]
		beanType := reflect.TypeOf(bean)
		log.Debug("  Trying %s", beanType)
		if beanType.String() == requiredTypeAsString {
			compatibleBeansIndexes = append(compatibleBeansIndexes, i)
			log.Debug("    Found %s as candidate for injection under type %s", getString(beanName, bean), beanType)
		}
	}

	if len(compatibleBeansIndexes) == 0 {
		return nil, errors.New("   no matches for requested type")
	}
	if len(compatibleBeansIndexes) > 1 {
		log.Critical("MULTIPLE INJECTION CANDIDATES")
		for i := 0; i < len(compatibleBeansIndexes); i++ {
			log.Critical(" - %s", initializedBeanNamesList[i])
		}
	}
	compatibleBeanIndex := compatibleBeansIndexes[0]
	return &initializedBeansList[compatibleBeanIndex], nil
}

func examiner(x interface{}) {
	t := reflect.TypeOf(x)
	examinerD(t, 1)
}

func examinerD(t reflect.Type, depth int) {
	fmt.Println(strings.Repeat("\t", depth), "Type is", t.Name(), "and kind is", t.Kind())
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		fmt.Println(strings.Repeat("\t", depth+1), "Contained type:")
		examinerD(t.Elem(), depth+1)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Println(strings.Repeat("\t", depth+1), "Field", i+1, "name is", f.Name, "type is", f.Type.Name(), "and kind is", f.Type.Kind())
			if f.Tag != "" {
				fmt.Println(strings.Repeat("\t", depth+2), "Tag is", f.Tag)
				fmt.Println(strings.Repeat("\t", depth+2), "tag1 is", f.Tag.Get("tag1"), "tag2 is", f.Tag.Get("tag2"))
			}
		}
	}
}

//func LoadConfig(configFileName string, configStruct interface{}) {
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Println("LoadConfig.Recovered: ", r)
//		}
//	}()
//	conf, err := toml.LoadFile(configFileName)
//	if err == nil {
//		v := reflect.ValueOf(configStruct)
//		typeOfS := v.Elem().Type()
//		sectionName := getTypeName(configStruct)
//		for i := 0; i < v.Elem().NumField(); i++ {
//			if v.Elem().Field(i).CanInterface() {
//				kName := conf.Get(sectionName + "." + typeOfS.Field(i).Name)
//				kValue := reflect.ValueOf(kName)
//				if (kValue.IsValid()) {
//					v.Elem().Field(i).Set(kValue.Convert(typeOfS.Field(i).Type))
//				}
//			}
//		}
//	} else {
//		fmt.Println("LoadConfig.Error: " + err.Error())
//	}
//}
