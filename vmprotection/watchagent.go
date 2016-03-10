/*
===========================================================================
ORBIT VM PROTECTOR GPL Source Code
Copyright (C) 2015 Vasileios Anagnostopoulos.
This file is part of the ORBIT VM PROTECTOR Source Code (?ORBIT VM PROTECTOR Source Code?).
ORBIT VM PROTECTOR Source Code is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
ORBIT VM PROTECTOR Source Code is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with ORBIT VM PROTECTOR Source Code.  If not, see <http://www.gnu.org/licenses/>.
In addition, the ORBIT VM PROTECTOR Source Code is also subject to certain additional terms. You should have received a copy of these additional terms immediately following the terms and conditions of the GNU General Public License which accompanied the Doom 3 Source Code.  If not, please request a copy in writing from id Software at the address below.
If you have questions concerning this license or the applicable additional terms, you may contact in writing Vasileios Anagnostopoulos, Campani 3 Street, Athens Greece, POBOX 11252.
===========================================================================
*/
// watchagent
package vmprotection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/fithisux/orbit-dc-protector/utilities"

	"github.com/oleiade/lane"
)

const (
	WATCH_RETRIES     = 3
	MongoTimeout      = 60
	AuthDatabase      = "orbitgoer"
	Reservoircapacity = 5
)

var servermutex sync.Mutex

type VMdata struct {
	Servervms   map[string]int `json:"watchdog_servervms"`
	Serverepoch int            `json:"watchdog_serverepoch"`
}

type Watchmedata struct {
	utilities.OPConfig
	Vmdata VMdata `json:"watchdog_vmdata"`
}

type Watchagentdesc struct {
	Watchmedata
	Agentparked bool `json:"watchdog_agentparked"`
}

type Watchagent struct {
	Watchagentdesc
	Observed         []Watchmedata      `json:"watchdog_observed"`
	Ovpwatchers      []utilities.OPData `json:"watchdog_watchers"`
	ovpconfig        utilities.OVPconfig
	persistencylayer *utilities.PersistencyLayer
	watching         map[utilities.OPConfig]VMdata
	memberlistagent  *MemberlistAgent
	observer         *Observer
}

type OrbitError struct {
	Status bool
	Reason string
}

func CreateWatchAgent(json *utilities.ServerConfig) *Watchagent {
	watchagent := new(Watchagent)
	watchagent.ovpconfig = json.Ovpconfig
	watchagent.observer = CreateObserver()
	watchagent.Agentparked = true
	watchagent.watching = make(map[utilities.OPConfig]VMdata)
	watchagent.persistencylayer = utilities.CreatePersistencyLayer(&json.Dbconfig)
	opdata := watchagent.persistencylayer.Initialize(&json.Opconfig)
	watchagent.OPConfig = opdata.OPConfig
	watchagent.Vmdata.Serverepoch = opdata.Epoch
	watchagent.Vmdata.Servervms = make(map[string]int)
	go watchagent.reportVMEvents()
	go watchagent.reportHostevents()
	watchagent.memberlistagent = CreateMemberlistAgent(opdata, watchagent.observer)
	fmt.Println("ok1")
	ticker := time.NewTicker(watchagent.ovpconfig.Refreshattempts.Timeout)
	go func() {
		for {
			select {
			case <-ticker.C:
				watchagent.Refreshwatchers()
				ticker = time.NewTicker(watchagent.ovpconfig.Refreshattempts.Timeout)
			}
		}
	}()

	fmt.Println("ok2")
	return watchagent
}

func (watchagent *Watchagent) reportVMEvents() {
	fmt.Println("Started reportvmevents")
	go Runme()
	fmt.Println("listening on tokenchan")
	for x := range Tokenchan {
		fmt.Println("got one")
		watchagent.Vmdata.Servervms = x.Vmlista
		if x.Status == 2 {
			MakeKnown(x.Vmuuid)
			vmdetection := new(utilities.VMDetection)
			vmdetection.Reporter_ovip = watchagent.Ovip
			vmdetection.Reporter_dcid = watchagent.Dcid
			vmdetection.Dcid = watchagent.Dcid
			vmdetection.Breakage = false
			vmdetection.Epoch = watchagent.Vmdata.Serverepoch
			vmdetection.Ovip = watchagent.Ovip
			vmdetection.Timestamp = time.Now()
			vmdetection.Vmid = []string{x.Vmuuid}
			watchagent.persistencylayer.InsertVMDetection(vmdetection)
		}
		fmt.Printf("Lista length for broadcasting == %d\n", len(x.Vmlista))
		servermutex.Lock()
		watchers := watchagent.Ovpwatchers
		servermutex.Unlock()
		if watchers != nil {
			q := BroadcastUpdate(&watchagent.Watchmedata, watchers) //TODO, do something on error
			fmt.Printf("Broadcasted to %d\n", q.Size())
		}
	}
}

func (watchagent *Watchagent) reportHostevents() {
	fmt.Println("Started reporthostevents")
	for nodevent := range watchagent.observer.Notifier {
		fmt.Println("hostevent " + strconv.Itoa(int(nodevent.Mesgtype)))
		if nodevent.Mesgtype == NOTIFY_LEAVE {
			fmt.Println("LEFT " + nodevent.Name)
			sd := new(utilities.OPData)

			//try to unmarshal from  name
			if err := json.Unmarshal([]byte(nodevent.Name), sd); err != nil {
				panic(err.Error)
			}

			//pass through if not running
			if temp := watchagent.isRunning(); temp != nil {
				return
			}

			announce := false
			servermutex.Lock()
			vmdata, ok := watchagent.watching[sd.OPConfig]
			if ok {
				if sd.Epoch < vmdata.Serverepoch { //stray previous detection but missed deregister
					panic("stray previous detection but missed deregister")
				} else if sd.Epoch > vmdata.Serverepoch {
					panic("left from future but missed register")
				} else {
					delete(watchagent.watching, sd.OPConfig)
					watchagent.recreateObservers()
					announce = true
				}
			}
			servermutex.Unlock()

			if announce {

				vmnames := make([]string, len(vmdata.Servervms))
				indy := 0
				for vmuuid := range vmdata.Servervms {
					_ = MakeKnown(vmuuid)
					vmnames[indy] = vmuuid
					indy++
				}
				vmdetection := new(utilities.VMDetection)
				vmdetection.Reporter_ovip = watchagent.Ovip
				vmdetection.Reporter_dcid = watchagent.Dcid
				vmdetection.Breakage = true
				vmdetection.Dcid = sd.Dcid
				vmdetection.Epoch = sd.Epoch
				vmdetection.Ovip = sd.Ovip
				vmdetection.Timestamp = time.Now()
				vmdetection.Vmid = vmnames
				watchagent.persistencylayer.InsertVMDetection(vmdetection)
			}
		}
	}
}

func (w *Watchagent) findWatchers() *lane.Queue {
	destinations := w.persistencylayer.GetOVPPeers(w.ovpconfig.Numofwatchers, &w.OPConfig)
	queue := BroadcastRegister(&w.Watchmedata, destinations)
	return queue
}

func (w *Watchagent) Join() *OrbitError {
	Agentparked := false
	servermutex.Lock()
	Agentparked = w.Agentparked
	servermutex.Unlock()

	if !Agentparked {
		return &OrbitError{false, "Is not parked"}
	}

	exposelist := w.persistencylayer.GetOVPPeers(w.ovpconfig.Numofwatchers, &w.OPConfig)

	w.memberlistagent.Join(exposelist)
	servermutex.Lock()
	w.Agentparked = false
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

func (w *Watchagent) Refreshwatchers() {
	/*
		temp := w.isRunning()
		if temp != nil {
			return
		}

		servermutex.Lock()
		thresh := len(w.Ovpwatchers)
		servermutex.Unlock()

		if thresh >= w.ovpconfig.Minwatchers {
			return
		}

		var q *lane.Queue
		found := false
		for i := 0; i < w.ovpconfig.Refreshattempts.Retries; i++ {
			q = w.findWatchers()
			found = (q.Size() >= w.ovpconfig.Minwatchers)
			if found {
				break
			}
		}

		if !found {
			return
		}

		servermutex.Lock()
		watchers := w.Ovpwatchers
		servermutex.Unlock()

		BroadcastWithdraw(&w.Watchmedata, watchers)

		watchers = make([]utilities.OPData, q.Size())
		for i := 0; i < len(watchers); i++ {
			watchers[i] = q.Pop().(utilities.OPData)
		}
		servermutex.Lock()
		w.Ovpwatchers = watchers
		servermutex.Unlock()
	*/
}

func (w *Watchagent) Start() *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}

	q := w.findWatchers()
	if q.Size() < w.ovpconfig.Minwatchers {
		return &OrbitError{false, "Not watched"}
	}

	watchers := make([]utilities.OPData, q.Size())
	for i := 0; i < len(watchers); i++ {
		watchers[i] = q.Pop().(utilities.OPData)
	}
	servermutex.Lock()
	w.Ovpwatchers = watchers
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

func (w *Watchagent) Stop() *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}
	var ovpwatchers []utilities.OPData
	servermutex.Lock()
	ovpwatchers = w.Ovpwatchers
	servermutex.Unlock()
	_ = BroadcastWithdraw(&w.Watchmedata, ovpwatchers)
	servermutex.Lock()
	w.Agentparked = false
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

func (w *Watchagent) isRunning() *OrbitError {
	Agentparked := true
	servermutex.Lock()
	Agentparked = w.Agentparked
	servermutex.Unlock()

	if Agentparked {
		return &OrbitError{false, "Is parked"}
	} else {
		return nil
	}
}

func (w *Watchagent) Register(wd *Watchmedata) *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}
	fmt.Println("Try to register")
	servermutex.Lock()
	vmdata, ok := w.watching[wd.OPConfig]
	if !ok {
		w.watching[wd.OPConfig] = wd.Vmdata
		fmt.Println("Successful registration")
		w.recreateObservers()
	} else {
		if wd.Vmdata.Serverepoch != vmdata.Serverepoch {
			panic("Register already registered with different epoch")
		} else {
			fmt.Println("Reregistration")
		}
	}
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

func (w *Watchagent) Withdraw(wd *Watchmedata) *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}
	fmt.Println("Try to withdraw")
	servermutex.Lock()
	vmdata, ok := w.watching[wd.OPConfig]
	if ok {
		if wd.Vmdata.Serverepoch != vmdata.Serverepoch {
			panic("Withdraw already registered with different epoch")
		} else {
			delete(w.watching, wd.OPConfig)
			w.recreateObservers()
			fmt.Println("Successful withdraw")
		}
	}
	servermutex.Unlock()

	if !ok {
		return &OrbitError{false, fmt.Sprintf("not found %+v", wd.OPConfig)}
	} else {
		return &OrbitError{true, ""}
	}
}

func (w *Watchagent) Update(wd *Watchmedata) *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}
	fmt.Println("Try to update")

	servermutex.Lock()
	vmdata, ok := w.watching[wd.OPConfig]
	if ok {
		if wd.Vmdata.Serverepoch != vmdata.Serverepoch {
			panic("Update already registered with different epoch")
		} else {
			w.watching[wd.OPConfig] = wd.Vmdata
			w.recreateObservers()
			fmt.Println("Successful update")
		}
	}
	servermutex.Unlock()

	if !ok {
		return &OrbitError{false, fmt.Sprintf("not found %s", wd.OPConfig)}
	} else {
		return &OrbitError{true, ""}
	}
}

func (w *Watchagent) recreateObservers() {
	w.Observed = make([]Watchmedata, len(w.watching))
	index := 0
	for k, v := range w.watching {
		w.Observed[index].OPConfig = k
		w.Observed[index].Vmdata = v
		index++
	}
}

func MakeKnown(vmuuid string) *OrbitError {

	fmt.Println("makeknown for " + vmuuid)
	cmd := exec.Command("./failover", vmuuid)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Notified? " + err.Error())
		return &OrbitError{false, err.Error()}
	} else {
		return &OrbitError{true, "Notified"}
	}
}
