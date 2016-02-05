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
