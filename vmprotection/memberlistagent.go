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

func CreateMemberlistAgent(ovpexpose *utilities.OVPExpose) *MemberlistAgent {
	ma := new(MemberlistAgent)
	fmt.Println("c1")
	c := memberlist.DefaultLocalConfig()
	fmt.Println("c3")
	c.Name = ovpexpose.Name()
	fmt.Println("c4")
	c.BindAddr = ovpexpose.Ovip
	c.BindPort = ovpexpose.Serfport
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

func (ma *MemberlistAgent) Join(exposelist []utilities.OVPExpose) {
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
