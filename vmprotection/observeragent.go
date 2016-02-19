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
	"fmt"

	"github.com/hashicorp/memberlist"
)

const (
	NOTIFY_JOIN = iota
	NOTIFY_LEAVE
	NOTIFY_UPDATE
)

type Observermesg struct {
	Name     string
	Mesgtype int
}

func CreateObserver() *Observer {
	o := new(Observer)
	o.Notifier = make(chan Observermesg)
	return o
}

type Observer struct {
	Notifier chan Observermesg
}

func (o *Observer) NotifyJoin(n *memberlist.Node) {
	fmt.Println("notifyjoin1 " + n.Name)
	if o.Notifier == nil {
		panic("NOnnotifier")
	}
	o.Notifier <- Observermesg{n.Name, NOTIFY_JOIN}
	fmt.Println("notifyjoin2 " + n.Name)
}

func (o *Observer) NotifyLeave(n *memberlist.Node) {
	fmt.Println("notifyleave " + n.Name)
	o.Notifier <- Observermesg{n.Name, NOTIFY_LEAVE}
}

func (o *Observer) NotifyUpdate(n *memberlist.Node) {
	fmt.Println("notifyupdate " + n.Name)
	o.Notifier <- Observermesg{n.Name, NOTIFY_UPDATE}
}
