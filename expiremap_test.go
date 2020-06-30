package expiremap

import (
	"path/filepath"
	"runtime"
	"strings"
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
	p := t.Logf
	e := t.Errorf
	var v interface{}
	var exist bool

	shouldExist := func() {
		_, li, fu := getCallerInfo(1)
		if false == exist {
			e("%v, Line %d, %s() - should exist but not", time.Now(), li, fu)
			return
		}
		p("%v, Line %d, %s() - check existance OK, value: %v", time.Now(), li, fu, v)
	}

	shouldNotExist := func() {
		_, li, fu := getCallerInfo(1)
		if exist {
			e("%v, Line %d, %s() - should not exist but does", time.Now(), li, fu)
			return
		}
		p("%v, Line %d, %s() - check existance OK", time.Now(), li, fu)
	}

	_, li, fu := getCallerInfo(0)
	p("%v, Line %d, %s() - test starts", time.Now(), li+1, fu)

	// create expiremap
	exp := 5 * time.Second
	m := New(exp)

	if m.Expiration() != exp {
		e("expiration not equal: %v", m.Expiration())
		m = nil
		time.Sleep(time.Second)
		runtime.GC()
		return
	}

	m.Store("1", time.Now())
	time.Sleep(time.Second)
	v, exist = m.Load("1")
	shouldExist()

	// extend it
	time.Sleep(time.Second)
	m.Store("1", time.Now())
	time.Sleep(exp - time.Second)
	v, exist = m.Load("1")
	shouldExist()

	time.Sleep(2 * time.Second)
	v, exist = m.Load("1")
	shouldNotExist()

	m = nil
	runtime.GC()
	time.Sleep(time.Second)
}

func TestInvalidExpiration(t *testing.T) {
	e := t.Errorf
	if m := New(0); m.Expiration() <= 0 {
		e("invalid expiration: %v", m.Expiration())
		return
	}
	runtime.GC()
	return
}
