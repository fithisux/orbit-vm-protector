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
package businesslogic

import (
	"fmt"
	"sync"
	//	"strconv"
	"github.com/fithisux/orbit-dc-protector/utilities"
	"github.com/oleiade/lane"
	//	"time"
	"encoding/json"
	"strconv"
)

const (
	NOTIFY_JOIN = iota
	NOTIFY_LEAVE
	NOTIFY_UPDATE
)

const (
	WATCH_RETRIES     = 3
	MongoTimeout      = 60
	AuthDatabase      = "orbitgoer"
	Reservoircapacity = 5
)

var servermutex sync.Mutex

type Watchagentdesc struct {
	Expose      utilities.OVPExpose   `json:"watchdog_expose"`
	Agentparked bool                  `json:"watchdog_agentparked"`
	Agentepoch  int                   `json:"watchdog_agentepoch"`
	Watchers    []utilities.OVPExpose `json:"watchdog_watchers"`
	Watched     map[string]Watchdesc  `json:"watchdog_watched"`
}

type Watchdesc struct {
	Watched_expose utilities.OVPExpose `json:"watched_expose"`
	Watched_vmdata VMdata              `json:"watched_vmdata"`
}

type Watchagent struct {
	Agentparked bool
	Serverconf  *utilities.ServerConfig
	pl          *utilities.PersistencyLayer
	watched     map[utilities.OVPExpose]VMdata
	ma          *MemberlistAgent
	Observeme   *Observer
	Watchers    []utilities.OVPExpose
}

type OrbitError struct {
	Status bool
	Reason string
}

func CreateWatchAgent(json *utilities.ServerConfig) *Watchagent {
	w := new(Watchagent)
	w.Serverconf = json
	w.Agentparked = true
	w.watched = make(map[utilities.OVPExpose]VMdata)
	w.pl = utilities.CreatePersistencyOVP(&json.Exposeconfig, &json.Dbconfig)
	vmdata := new(VMdata)
	vmdata.Serverepoch = w.pl.Ovpdata.Epoch
	w.watched[json.Exposeconfig.Ovpexpose] = *vmdata
	w.Observeme = CreateObserver()
	go ReportHostevents(w)
	w.ma = CreateMemberlistAgent(w)
	go ReportVMEvents(w)
	return w
}

func ReportVMEvents(w *Watchagent) {
	fmt.Println("Started reportvmevents")
	go Runme()
	fmt.Println("listening on tokenchan")
	for x := range Tokenchan {
		fmt.Println("got one")
		vmdata := VMdata{w.pl.Ovpdata.Epoch, x.Vmlista}
		servermutex.Lock()
		w.watched[w.pl.Ovpdata.OVPExpose] = vmdata
		servermutex.Unlock()
		if x.Status == 2 {
			MakeKnown(x.Vmuuid, w.Serverconf)
		}
		fmt.Printf("Broadcasted to %d\n", len(x.Vmlista))
		wd := Watchmedata{w.pl.Ovpdata.OVPExpose, vmdata}
		servermutex.Lock()
		wwatchers := w.Watchers
		servermutex.Unlock()
		if wwatchers != nil {
			q := Broadcast(&wd, wwatchers) //TODO, do something on error
			fmt.Printf("Broadcasted to %d\n", q.Size())
		}
	}
}

func ReportHostevents(w *Watchagent) {
	fmt.Println("Started reporthostevents")
	for mesg := range w.Observeme.Notifier {
		fmt.Println("hostevent " + strconv.Itoa(mesg.Mesgtype))
		if mesg.Mesgtype == NOTIFY_LEAVE {
			fmt.Println("LEFT " + mesg.Name)

			sd := new(utilities.OVPExpose)
			err := json.Unmarshal([]byte(mesg.Name), sd)

			if err != nil {
				panic(err.Error)
			}

			if temp := w.isRunning(); temp != nil {
				return
			}

			ok := false
			var vmdata VMdata
			servermutex.Lock()
			vmdata, ok = w.watched[*sd]
			if ok {
				delete(w.watched, *sd)
			}
			servermutex.Unlock()

			if ok {
				w.pl.Changeweight(-1)
				w.pl.Makefailed(sd)
				for vmuuid := range vmdata.Servervms {
					_ = MakeKnown(vmuuid, w.Serverconf)
				}

			}
		}
	}
}

func (w *Watchagent) findWatchers(wd *Watchmedata, bound int) *lane.Queue {

	destinations := w.pl.GetOVPPeers(bound)
	queue := Broadcast(wd, destinations)
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

	w.ma.Join(w.pl)
	servermutex.Lock()
	w.Agentparked = false
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

func (w *Watchagent) Start() *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}

	wd := new(Watchmedata)
	wd.Expose = w.pl.Ovpdata.OVPExpose
	servermutex.Lock()
	wd.Serverdata = w.watched[wd.Expose]
	servermutex.Unlock()
	q := w.findWatchers(wd, w.Serverconf.Numofwatchers)
	if q.Size() == 0 {
		return &OrbitError{false, "Not watched"}
	}

	wwatchers := make([]utilities.OVPExpose, q.Size())
	for i := 0; i < len(wwatchers); i++ {
		wwatchers[i] = q.Pop().(utilities.OVPExpose)
	}
	servermutex.Lock()
	w.Watchers = wwatchers
	servermutex.Unlock()
	return &OrbitError{true, ""}
}

type VMdata struct {
	Serverepoch int
	Servervms   map[string]int
}

type Watchmedata struct {
	Expose     utilities.OVPExpose
	Serverdata VMdata
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

func (w *Watchagent) Watch(wd *Watchmedata) *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}

	fmt.Println("Try to add watchable")
	ok := false
	var vmdata VMdata
	servermutex.Lock()
	vmdata, ok = w.watched[wd.Expose]
	if !ok || wd.Serverdata.Serverepoch >= vmdata.Serverepoch {
		fmt.Println("Add watching " + wd.Expose.Name())
		w.watched[wd.Expose] = wd.Serverdata
	}
	servermutex.Unlock()
	w.pl.Changeweight(1)
	return &OrbitError{true, ""}
}

func (w *Watchagent) Unwatch(wd *Watchmedata) *OrbitError {
	temp := w.isRunning()
	if temp != nil {
		return temp
	}

	ok := false
	var vmdata VMdata
	servermutex.Lock()
	vmdata, ok = w.watched[wd.Expose]
	if ok && wd.Serverdata.Serverepoch >= vmdata.Serverepoch {
		delete(w.watched, wd.Expose)
	}
	servermutex.Unlock()
	w.pl.Changeweight(-1)

	if !ok {
		return &OrbitError{false, fmt.Sprintf("not found %s", wd.Expose)}
	} else {
		return &OrbitError{true, ""}
	}
}

func (w *Watchagent) Describe() Watchagentdesc {
	var wad Watchagentdesc
	servermutex.Lock()
	wad.Expose = w.pl.Ovpdata.OVPExpose
	wad.Agentepoch = w.pl.Ovpdata.Epoch
	wad.Agentparked = w.Agentparked
	wad.Watchers = w.Watchers
	wad.Watched = make(map[string]Watchdesc)
	for k, v := range w.watched {
		wad.Watched[k.Name()] = Watchdesc{k, v}
	}
	servermutex.Unlock()
	return wad
}
