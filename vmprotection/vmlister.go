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
// jsonutil.go

package vmprotection

// #cgo pkg-config: libvirt
// #include <stdio.h>
// extern int main1();
import "C"
import (
	//	"unsafe"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	VIR_DOMAIN_EVENT_DEFINED = iota
	VIR_DOMAIN_EVENT_UNDEFINED
	VIR_DOMAIN_EVENT_STARTED
	VIR_DOMAIN_EVENT_SUSPENDED
	VIR_DOMAIN_EVENT_RESUMED
	VIR_DOMAIN_EVENT_STOPPED
	VIR_DOMAIN_EVENT_SHUTDOWN
	VIR_DOMAIN_EVENT_PMSUSPENDED
	VIR_DOMAIN_EVENT_CRASHED
	VIR_DOMAIN_EVENT_LAST
)

type Listtoken struct {
	Status  int
	Vmuuid  string
	Vmlista map[string]int
}

var Tokenchan chan Listtoken = make(chan Listtoken)

//export myReportEvent
func myReportEvent(vir_dom_name *C.char, vir_dom_id C.uint, vir_event C.int, vir_detail C.int) {
	vmid := int(vir_dom_id)
	vmuuid := C.GoString(vir_dom_name)
	fmt.Printf("Reporting ID = %d and UUID = %s\n", vmid, vmuuid)

	eventtype := int(vir_event)
	eventdetail := int(vir_detail)
	fmt.Println("Event is " + strconv.Itoa(eventtype))
	fmt.Println("Detail is " + strconv.Itoa(eventdetail))
	switch eventtype {
	case VIR_DOMAIN_EVENT_CRASHED:
		{
			if eventdetail == 2 || eventdetail == 5{
				fmt.Println("Crashed -1")
				delete(vmlista, vmuuid)
				Tokenchan <- Listtoken{2, vmuuid, vmlista}				
			}			
		}
	case VIR_DOMAIN_EVENT_SHUTDOWN:
		{
			fmt.Println("Shutdown -1")
			delete(vmlista, vmuuid)
			Tokenchan <- Listtoken{1, vmuuid, vmlista}
		}
	case VIR_DOMAIN_EVENT_STARTED:
		{
			fmt.Println("Started +1")
			vmlista[vmuuid] = vmid
			Tokenchan <- Listtoken{0, vmuuid, vmlista}
		}
	}
}

var vmlista map[string]int = make(map[string]int)

//export myReportID
func myReportID(vir_dom_name *C.char, vir_dom_id C.uint) {
	vmid := int(vir_dom_id)
	vmuuid := C.GoString(vir_dom_name)
	fmt.Printf("Reporting ID = %d and UUID = %s\n", vmid, vmuuid)
	vmlista[vmuuid] = vmid
}

//export myReportStraight
func myReportStraight() {
	fmt.Println("to upper level " + strconv.Itoa(len(vmlista)))
	Tokenchan <- Listtoken{-1, "", vmlista}
}

func cleanup() {
	fmt.Println("cleanup")
}

func Runme() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	defer close(Tokenchan)
	if res := C.main1(); res == -1 {
		fmt.Println("Ooops went wrong")
		os.Exit(-1)
	}
}
