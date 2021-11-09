package evq

import "fmt"

type CoLock struct {
	name string
	cs   map[interface{}]chan struct{}
}

func NewLock(name string) *CoLock {
	return &CoLock{
		name: name,
		cs:   map[interface{}]chan struct{}{},
	}
}

func (l *CoLock) Lock(key interface{}) chan struct{} {
	c, ok := l.cs[key]
	if ok {
		<-c
		return l.Lock(key)
	} else {
		c = make(chan struct{})
		l.cs[key] = c
		return c
	}
}

func (l *CoLock) UnLock(key interface{}, c chan struct{}) {
	ch := l.cs[key]
	if ch != c {
		panic(fmt.Sprintf("CoLock.UnLock name=%s, key=%s", l.name, key))
	}

	delete(l.cs, key)
	close(ch)
}

func (l *CoLock) WaitOrLock(key interface{}) chan struct{} {
	c, ok := l.cs[key]
	if ok {
		<-c
		return nil
	} else {
		c = make(chan struct{})
		l.cs[key] = c
		return c
	}
}
