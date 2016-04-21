package libkademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	RouteTable  []K_Buckets			// 0 based 
	RTManagerChan   chan Contact
}

 


func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID

	// TODO: Initialize other state here as you add functionality.
	 !!!initial 160 lists, go updatethread,initial channel
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}
	s.HandleHTTP(rpc.DefaultRPCPath+hostname+port,
		rpc.DefaultDebugPath+hostname+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// handle update
	// go k.RouteTableUpdateHandler()

	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ = net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	return k
}

func NewKademlia(laddr string) *Kademlia {
	return NewKademliaWithId(laddr, NewRandomID())
}

type ContactNotFoundError struct {
	id  ID
	msg string
}

func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

// find the left's bucket num containing right

func FindBucketNum(left ID, right ID) int {
	res := left.Xor(right)
	dig := b - 1
	for i := len(res); i >= 0; i-- {
		if(res[i] != 0){
			dig = i
			break
		}
	}

	for j := 7; j >= 0; j-- {
		if(res[dig] >> uint8(j) & 0x1 != 0){
			return (dig - 1) * 8 + j               //bot, num of list
		}
	} 
	return 161
}

// x
func (k *Kademlia) UpdateRouteTable(c *Contact) {
	num := FindBucketNum(k.NodeID, c.NodeID) 
	l := k.RouteTable[num]								// list should contain c

	_, erro := k.FindContact(c.NodeID)
	if erro == nil{
		l.MoveToTail(c)
	} else {
		if !l.CheckFull() {
			l.AddTail(c)
		} else {
			head := l.GetHead()
			_, erro := k.DoPing(head.Host, head.Port)
			if erro != nil {
				l.RemoveHead()
				l.AddTail(c)
			}
		} 
	}
}

// one thread running Ro..Handler for thread safe
func (k *Kademlia) RouteTableUpdateHandler() {
	for {
		select {
			case c := <- k.RTManagerChan: 
				k.UpdateRouteTable(&c)
		}
	}
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	} else {
		num := FindBucketNum(k.NodeID, nodeId)			//find number of list
		l := k.RouteTable[num].bucket
		for e := l.Front(); e != nil; e = e.Next() {
			if e.Value.(* Contact).NodeID.Equals(nodeId) {
				return e.Value.(* Contact), nil
			}

		}
	}
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}

type CommandFailed struct {
	msg string
}

func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}

func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement
	return nil, &CommandFailed{
		"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	// TODO: Implement
	return &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	return nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	return nil, nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) LocalFindValue(searchKey ID) ([]byte, error) {
	// TODO: Implement
	return []byte(""), &CommandFailed{"Not implemented"}
}

// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	return nil, &CommandFailed{"Not implemented"}
}
func (k *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	return nil, &CommandFailed{"Not implemented"}
}

// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}
