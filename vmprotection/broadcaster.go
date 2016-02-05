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
// restclient.go
package vmprotection

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/fithisux/orbit-dc-protector/utilities"
	"github.com/jmcvetta/napping"
	"github.com/oleiade/lane"
)

func BroadcastUpdate(watchmetadata *Watchmedata, destinations []utilities.OPData) *lane.Queue {
	var wg sync.WaitGroup
	wg.Add(len(destinations))
	var queue *lane.Queue = lane.NewQueue()

	for i := 0; i < len(destinations); i++ {
		go callPeer(watchmetadata, &destinations[i], queue, "update", &wg)
	}

	wg.Wait()
	return queue
}

func BroadcastRegister(watchmetadata *Watchmedata, destinations []utilities.OPData) *lane.Queue {
	var wg sync.WaitGroup
	wg.Add(len(destinations))
	var queue *lane.Queue = lane.NewQueue()
	for i := 0; i < len(destinations); i++ {
		go callPeer(watchmetadata, &destinations[i], queue, "register", &wg)
	}
	wg.Wait()
	return queue
}

func BroadcastWithdraw(watchmetadata *Watchmedata, destinations []utilities.OPData) *lane.Queue {
	var wg sync.WaitGroup
	wg.Add(len(destinations))
	var queue *lane.Queue = lane.NewQueue()
	for i := 0; i < len(destinations); i++ {
		go callPeer(watchmetadata, &destinations[i], queue, "withdraw", &wg)
	}
	wg.Wait()
	return queue
}

func callPeer(watchmetadata *Watchmedata, destination *utilities.OPData, queue *lane.Queue, verb string, wg *sync.WaitGroup) {
	defer wg.Done()

	target := "http://" + destination.Ovip + ":" + strconv.Itoa(destination.Announceport) + "/watchdog"
	url := target + "/" + verb
	fmt.Println("Instructed as " + url)
	e := struct {
		Message string
		Errors  []struct {
			Resource string
			Field    string
			Code     string
		}
	}{}

	var j map[string]interface{}

	s := napping.Session{}

	resp, err := s.Post(url, watchmetadata, &j, &e)

	if err != nil {
		log.Printf("failed to " + verb + " on " + target + " because " + err.Error())
		return
	}

	if resp.Status() != 200 {
		log.Printf("failed to " + verb + " on " + target + " because " + strconv.Itoa(resp.Status()) + "..." + e.Message)
		return
	}

	if verb == "withdraw" || verb == "update" {
		source := watchmetadata.OPConfig
		if j["status"].(bool) {
			queue.Enqueue(destination)
		} else {
			mesg := fmt.Sprintf("Not "+verb+" %+v from %+v", source, destination)
			panic(mesg)
		}
	}
	queue.Enqueue(*destination)
}
