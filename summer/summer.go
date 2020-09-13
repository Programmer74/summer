package summer

import (
	"bufio"
	"errors"
	log "github.com/jeanphorn/log4go"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const SummerBeanTag = "summer"
const SummerPropertyTag = "summer.property"

var uninitializedBeans = make(map[string]interface{})

var initializedBeansList = make([]interface{}, 0)
var initializedBeanNamesList = make([]string, 0)

var beanTypeAliasToBeanNameMap = make(map[string]string)

var propertiesMap = make(map[string]string)

var beanDependenciesVertexList = make([]string, 0)

func GetPropertyValue(key string) (string, bool) {
	value, found := propertiesMap[key]
	return value, found
}

func RegisterBean(beanName string, bean interface{}) {
	uninitializedBeans[beanName] = bean
}

func RegisterBeanWithTypeAlias(beanName string, bean interface{}, beanType string) {
	RegisterBean(beanName, bean)
	beanTypeAliasToBeanNameMap[beanType] = beanName
}

func GetBean(beanName string) *interface{} {
	for i := 0; i < len(initializedBeansList); i++ {
		if initializedBeanNamesList[i] == beanName {
			return &initializedBeansList[i]
		}
	}
	log.Warn("no processed bean found for name", beanName)
	return nil
}

//use with https://yuml.me/diagram/scruffy/class/draw for best possible value
func PrintDependencyGraphVertex() {
	str := "\nDependency graph:"
	for i := 0; i < len(beanDependenciesVertexList); i++ {
		str = str + "\n" + beanDependenciesVertexList[i]
	}
	log.Warn(str)
}

func ParseProperties(path string) {
	//todo: full format coverage?
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		arr := strings.Split(line, "=")
		propertyKey := arr[0]
		propertyVal := arr[1]
		propertiesMap[propertyKey] = propertyVal
	}
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
				_, ok := f.Tag.Lookup(SummerBeanTag)
				if ok {
					annotatedDependenciesCount++
				}
				_, ok = f.Tag.Lookup(SummerPropertyTag)
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

	for fieldIndex := 0; fieldIndex < sourceType.NumField(); fieldIndex++ {
		sourceField := sourceType.Field(fieldIndex)
		if sourceField.Tag != "" {
			qualifier, beanTagFound := sourceField.Tag.Lookup(SummerBeanTag)
			if beanTagFound {
				summerTotalDependenciesCount, summerInjectedDependenciesCount =
					tryFillBeans(beanName, sourceField, summerTotalDependenciesCount, xCopy, fieldIndex, qualifier, summerInjectedDependenciesCount)
			}
			requiredPropertyParam, propertyTagFound := sourceField.Tag.Lookup(SummerPropertyTag)
			if propertyTagFound {
				tryFillProperties(sourceField, xCopy, fieldIndex, requiredPropertyParam)
			}
		}
	}

	unprocessedDependenciesLeft := summerTotalDependenciesCount - summerInjectedDependenciesCount
	return unprocessedDependenciesLeft, xCopy
}

func tryFillBeans(beanName string, sourceField reflect.StructField, summerTotalDependenciesCount int, xCopy interface{}, fieldIndex int, qualifier string, summerInjectedDependenciesCount int) (int, int) {
	log.Debug("Injectable 'summer' bean field %s", sourceField)
	summerTotalDependenciesCount++

	targetField := reflect.ValueOf(xCopy).Elem().Field(fieldIndex)
	if targetField.IsValid() {
		if targetField.Kind() != reflect.Ptr {
			panic("field annotated by 'summer' should be pointers")
		}
		if targetField.IsNil() {
			log.Debug("Field %s is not set", targetField)

			applicableBean, applicableBeanName, err := getProcessedBean(qualifier)

			if err == nil {
				log.Debug("Field %s value found : %s", targetField, applicableBean)
				if !targetField.CanSet() {
					log.Critical("CANNOT SET FIELD %s", targetField)
				}

				if targetField.Type().String() == "*interface {}" {
					targetField.Set(reflect.ValueOf(applicableBean).Convert(targetField.Type()))
					summerInjectedDependenciesCount++
					beanDependenciesVertexList = append(beanDependenciesVertexList, "["+beanName+"]->["+applicableBeanName+"]")
				} else {
					panic("field setter injection not implemented yet")
				}

			} else {
				log.Debug("Field %s value not found by now", targetField)
			}
		} else {
			log.Debug("Field %s already set", targetField)
			summerInjectedDependenciesCount++
		}
	} else {
		log.Error("For some reason, %s is not valid", targetField)
	}
	return summerTotalDependenciesCount, summerInjectedDependenciesCount
}

func tryFillProperties(sourceField reflect.StructField, xCopy interface{}, fieldIndex int, requiredPropertyParam string) {
	log.Debug("Injectable 'summer' property field %s", sourceField)

	arr := strings.Split(requiredPropertyParam, "|")
	propertyKey := arr[0]

	targetField := reflect.ValueOf(xCopy).Elem().Field(fieldIndex)
	if targetField.IsValid() {

		propertyValue, ok := propertiesMap[propertyKey]

		if !ok {
			if len(arr) == 1 {
				panic("property value not found and no default value specified")
			} else {
				propertyValue = arr[1]
			}
		}
		if targetField.Type().Name() == "int" {
			valueAsInt, err := strconv.Atoi(propertyValue)
			if err != nil {
				log.Error(err)
			}
			targetField.Set(reflect.ValueOf(valueAsInt))
		} else {
			//todo: cover other types as well
			targetField.Set(reflect.ValueOf(propertyValue))
		}

	} else {
		log.Error("For some reason, %s is not valid", targetField)
	}
}

func getProcessedBean(qualifier string) (*interface{}, string, error) {
	if strings.HasPrefix(qualifier, "*") {
		return getProcessedBeanByType(qualifier)
	}
	return getProcessedBeanByName(qualifier)
}

func getProcessedBeanByName(requiredBeanName string) (*interface{}, string, error) {
	log.Debug("  Asked to find '%s'", requiredBeanName)
	beanIndex := getProcessedBeanIndex(requiredBeanName)
	if beanIndex >= 0 {
		return &initializedBeansList[beanIndex], requiredBeanName, nil
	}
	return nil, "", errors.New("no matches for requested name")
}

func getProcessedBeanByType(requiredTypeAsString string) (*interface{}, string, error) {
	log.Debug("  Asked to find '%s'", requiredTypeAsString)

	compatibleBeansIndexes := getCompatibleBeansIndexes(requiredTypeAsString)
	beanNameWithSpecifiedAlias, found := beanTypeAliasToBeanNameMap[requiredTypeAsString]
	if found {
		log.Debug("For %s there is a bean '%s' specified separately", requiredTypeAsString, beanNameWithSpecifiedAlias)
		beanIndex := getProcessedBeanIndex(beanNameWithSpecifiedAlias)
		if beanIndex >= 0 {
			compatibleBeansIndexes = append(compatibleBeansIndexes, beanIndex)
		}
	}

	if len(compatibleBeansIndexes) == 0 {
		return nil, "", errors.New("no matches for requested type")
	}
	if len(compatibleBeansIndexes) > 1 {
		log.Critical("MULTIPLE INJECTION CANDIDATES")
		for i := 0; i < len(compatibleBeansIndexes); i++ {
			compatibleBeanIndex := compatibleBeansIndexes[i]
			log.Critical(" - %s", initializedBeanNamesList[compatibleBeanIndex])
		}
		panic("multiple injection candidates; see logs above")
	}
	compatibleBeanIndex := compatibleBeansIndexes[0]
	return &initializedBeansList[compatibleBeanIndex], initializedBeanNamesList[compatibleBeanIndex], nil
}

func getCompatibleBeansIndexes(requiredTypeAsString string) []int {
	compatibleBeansIndexes := make([]int, 0)
	for i := 0; i < len(initializedBeansList); i++ {
		compatibleBeansIndexes = tryForCompatibility(i, requiredTypeAsString, compatibleBeansIndexes)
	}
	return compatibleBeansIndexes
}

func tryForCompatibility(beanIndex int, requiredTypeAsString string, compatibleBeansIndexes []int) []int {
	bean := initializedBeansList[beanIndex]
	beanName := initializedBeanNamesList[beanIndex]
	beanType := reflect.TypeOf(bean)
	log.Debug("  Trying %s", beanType)
	if beanType.String() == requiredTypeAsString {
		compatibleBeansIndexes = append(compatibleBeansIndexes, beanIndex)
		log.Debug("    Found %s as candidate for injection under type %s", getString(beanName, bean), beanType)
	}
	return compatibleBeansIndexes
}

func getProcessedBeanIndex(beanName string) int {
	for i := 0; i < len(initializedBeansList); i++ {
		if initializedBeanNamesList[i] == beanName {
			return i
		}
	}
	return -1
}
