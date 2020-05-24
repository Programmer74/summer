package summer

import (
	"errors"
	log "github.com/jeanphorn/log4go"
	"reflect"
)

const SUMMER_BEAN_TAG = "summer"

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
			annotatedDependenciesCount := getAnnotatedDependenciesCount(bean)
			if annotatedDependenciesCount == 0 {
				log.Debug("Bean %s has no summer dependencies", getString(beanName, bean))
				oneBeanInitialized = advanceUnprocessedBeanAsInitialized(beanName, bean)
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

func advanceUnprocessedBeanAsInitialized(beanName string, unprocessedBean interface{}) bool {

	sourceType := reflect.TypeOf(unprocessedBean)

	xc := reflect.New(sourceType)
	xc.Elem().Set(reflect.ValueOf(unprocessedBean))

	bean := xc.Interface().(interface{})

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
				_, ok := f.Tag.Lookup(SUMMER_BEAN_TAG)
				if ok {
					annotatedDependenciesCount++
				}
			}
		}
	}
	return annotatedDependenciesCount
}

func tryFillDependencies(beanName string) (int, interface{}) {
	summerTotalDependenciesCount := 0
	summerInjectedDependenciesCount := 0

	x := uninitializedBeans[beanName]
	sourceType := reflect.TypeOf(x)

	xc := reflect.New(sourceType)
	xc.Elem().Set(reflect.ValueOf(x))

	xCopy := xc.Interface().(interface{})

	for i := 0; i < sourceType.NumField(); i++ {
		sourceField := sourceType.Field(i)
		if sourceField.Tag != "" {
			requiredTypeAsString, ok := sourceField.Tag.Lookup(SUMMER_BEAN_TAG)
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

							if targetField.Type().String() == "*interface {}" {
								targetField.Set(reflect.ValueOf(applicableBean).Convert(targetField.Type()))
								summerInjectedDependenciesCount++
							} else {
								//setterMethodName := "SummerSet" + sourceField.Name
								panic("field setter injection not implemented yet")
							}

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
