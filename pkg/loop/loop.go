/*
Copyright 2022 Luis Pabon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package loop

import (
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Loop struct {
	funcs     []func() error
	teardown  []func()
	maxperiod int
	stop      chan bool
	wg        sync.WaitGroup
}

func NewLoop(maxperiod int) *Loop {
	rand.Seed(time.Now().UnixNano())
	return &Loop{
		funcs:     make([]func() error, 0),
		teardown:  make([]func(), 0),
		stop:      make(chan bool),
		maxperiod: maxperiod,
	}
}

func (l *Loop) Add(f func() error) {
	l.funcs = append(l.funcs, f)
}

func (l *Loop) AddTeardown(f func()) {
	l.teardown = append(l.teardown, f)
}

func (l *Loop) Stop() {
	var i int
	for i = 0; i < len(l.funcs); i++ {
		l.stop <- true
	}
}

func (l *Loop) Start() {
	l.wg.Add(len(l.funcs))
	for _, f := range l.funcs {
		go l.loop(f)
	}
}

func (l *Loop) Wait() {
	l.wg.Wait()

	for _, f := range l.teardown {
		f()
	}
}

func (l *Loop) loop(f func() error) {
	stop := false
	defer l.wg.Done()

	for !stop {
		r := rand.Intn(l.maxperiod * 1000)
		ticker := time.NewTimer(time.Duration(r) * time.Millisecond)

		err := f()
		if err != nil {
			logrus.Errorf("%v", err)
		}

		select {
		case <-ticker.C:
		case stop = <-l.stop:
		}
	}
}
