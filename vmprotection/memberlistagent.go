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
package vmprotection

import (
	"fmt"
	"strconv"

	"github.com/fithisux/orbit-dc-protector/utilities"
	"github.com/hashicorp/memberlist"
)

type MemberlistAgent struct {
	conf *memberlist.Config
	list *memberlist.Memberlist
	Ch   chan memberlist.NodeEvent
}

func CreateMemberlistAgent(opdata *utilities.OPData) *MemberlistAgent {
	ma := new(MemberlistAgent)
	fmt.Println("c1")
	c := memberlist.DefaultLocalConfig()
	fmt.Println("c3")
	c.Name = opdata.Name()
	fmt.Println("c4")
	c.BindAddr = opdata.Ovip
	c.BindPort = opdata.Serfport
	observer := new(memberlist.ChannelEventDelegate)
	ma.Ch = make(chan memberlist.NodeEvent)
	observer.Ch = ma.Ch
	c.Events = observer
	fmt.Println("c5")
	list, err := memberlist.Create(c)
	fmt.Println("c6")
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}
	ma.list = list
	ma.conf = c
	fmt.Println("MMBL created")
	return ma
}

func (ma *MemberlistAgent) Join(opdatalist []utilities.OPData) {
	if len(opdatalist) >= 1 {
		peerlist := make([]string, len(opdatalist))
		for i := 0; i < len(opdatalist); i++ {
			peerlist[i] = opdatalist[i].Ovip + ":" + strconv.Itoa(opdatalist[i].Serfport)
			fmt.Println("Join point " + peerlist[i])
		}
		_, err := ma.list.Join(peerlist)
		if err != nil {
			panic(err.Error())
		}
	} else {
		panic("Small join failure")
	}
}
