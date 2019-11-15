package expiremap

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func getCallerInfo(invokeLevel int) (fileName string, line int, funcName string) {
	funcName = "FILE"
	line = -1
	fileName = "FUNC"

	if invokeLevel <= 0 {
		invokeLevel = 2
	} else {
		invokeLevel++
	}

	pc, fileName, line, ok := runtime.Caller(invokeLevel)
	if ok {
		fileName = filepath.Base(fileName)
		funcName = runtime.FuncForPC(pc).Name()
		funcName = filepath.Ext(funcName)
		funcName = strings.TrimPrefix(funcName, ".")
		// funcName, _ = url.QueryUnescape(funcName)
	}
	// fmt.Println(reflect.TypeOf(pc), reflect.ValueOf(pc))
	return
}

func TestBasicFunctional(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go test(t, &wg)
	wg.Wait()

	runtime.GC()
	time.Sleep(8 * time.Second)
	runtime.GC()
	time.Sleep(2 * time.Second)
}

func test(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	p := t.Logf
	e := t.Errorf
	var v interface{}
	var exist bool

	shouldExist := func() {
		_, li, fu := getCallerInfo(1)
		if false == exist {
			e("%v, %d, %s() - should exist but not", time.Now(), li, fu)
			return
		}
		p("%v, %d, %s() - check existance OK, value: %v", time.Now(), li, fu, v)
	}

	shouldNotExist := func() {
		_, li, fu := getCallerInfo(1)
		if exist {
			e("%v, %d, %s() - should not exist but does", time.Now(), li, fu)
			return
		}
		p("%v, %d, %s() - check existance OK", time.Now(), li, fu)
	}

	// create expiremap
	m := New(5*time.Second, time.Second)

	m.Store("1", time.Now())
	time.Sleep(time.Second)
	v, exist = m.Load("1")
	shouldExist()

	time.Sleep(4 * time.Second)
	v, exist = m.Load("1")
	shouldNotExist()
}
