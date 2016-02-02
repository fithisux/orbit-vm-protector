// observeragent.go
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
}

type Observermesg struct {
	Name     string
	Mesgtype int
}

func CreateMemberlistAgent(watchagent *Watchagent) *MemberlistAgent {
	ma := new(MemberlistAgent)
	fmt.Println("c1")
	c := memberlist.DefaultLocalConfig()
	fmt.Println("c3")
	c.Name = watchagent.Ovpdata.OVPExpose.Name()
	fmt.Println("c4")
	c.BindAddr = watchagent.Ovpdata.OVPExpose.Ovip
	c.BindPort = watchagent.Ovpdata.OVPExpose.Serfport
	c.Events = watchagent.Observeme
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

func (ma *MemberlistAgent) Join(persistencylayer *utilities.PersistencyLayer, bound int, ovpdata *utilities.OVPData) {
	exposelist := persistencylayer.GetOVPPeers(bound, ovpdata)

	if len(exposelist) >= 1 {
		peerlist := make([]string, len(exposelist))
		for i := 0; i < len(exposelist); i++ {
			peerlist[i] = exposelist[i].Ovip + ":" + strconv.Itoa(exposelist[i].Serfport)
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
