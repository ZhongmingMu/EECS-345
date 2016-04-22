package libkademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	"fmt"
)

type KademliaRPC struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (k *KademliaRPC) Ping(ping PingMessage, pong *PongMessage) error {
	fmt.Println(k.kademlia.SelfContact.Port)
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID)
	// Specify the sender
	pong.Sender = k.kademlia.SelfContact
	// Update contact, etc
	k.kademlia.RTManagerChan <- ping.Sender
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (k *KademliaRPC) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.
	
	k.kademlia.DataStoreChan <- KVpair{req.Key, req.Value}
	k.kademlia.RTManagerChan <- req.Sender
	
	res.MsgID = CopyID(req.MsgID)
	res.Err = nil
	
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func FormatTrans(c *Contact) Contact{
	newcontact := new(Contact)
	newcontact.Host = c.Host
	newcontact.NodeID = c.NodeID
	newcontact.Port = c.Port
	
	return *newcontact
	
}

func (kk *Kademlia) findCloestNodes(nodeid ID) []Contact {
	nodes := ([]Contact{})
	
	closestnum := FindBucketNum(kk.NodeID, nodeid)
	count := 0
	diff := 1
	fmt.Printf("%d: closetnumber  %d\n", kk.SelfContact.Port, closestnum)
	for e := kk.RouteTable[closestnum].bucket.Front(); e != nil; e = e.Next() {
		fmt.Printf("%d: here count %d\n", kk.SelfContact.Port, count)
		fmt.Printf("%d: list length %d\n", kk.SelfContact.Port, kk.RouteTable[closestnum].bucket.Len())
		fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, FormatTrans(e.Value.(*Contact)).Port)
		//fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, nodes[0].Port)
		nodes = append(nodes, FormatTrans(e.Value.(*Contact)))	
		count = count + 1
	}
	for ; count < k; {
		if closestnum - diff >= 0 {
			for e := kk.RouteTable[closestnum - diff].bucket.Front(); e != nil; e = e.Next() {
				fmt.Printf("%d: here count %d", kk.SelfContact.Port, count)
				nodes = append(nodes, FormatTrans(e.Value.(*Contact)))	
				count = count + 1
				if(count >= k - 1) {
					break					
				}
			}
		}
		
		if(count >= k - 1) {
			break
		}
		
		if closestnum + diff < b{
			for e := kk.RouteTable[closestnum + diff].bucket.Front(); e != nil; e = e.Next() {
				nodes = append(nodes, FormatTrans(e.Value.(*Contact)))	
				count = count + 1
				if(count >= k - 1){
					break
				}
			}
		}
		diff = diff + 1
	}
	return nodes
}

func (k *KademliaRPC) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	nodeid := req.NodeID
	
	res.MsgID = CopyID(req.MsgID)
	res.Nodes = k.kademlia.findCloestNodes(nodeid)
	res.Err = nil
	
	k.kademlia.RTManagerChan <- req.Sender
	return nil
}
	

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error    
}

func (k *KademliaRPC) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	return nil
}

// For Project 3

type GetVDORequest struct {
	Sender Contact
	VdoID  ID
	MsgID  ID
}

type GetVDOResult struct {
	MsgID ID
	VDO   VanashingDataObject
}

func (k *KademliaRPC) GetVDO(req GetVDORequest, res *GetVDOResult) error {
	// TODO: Implement.
	return nil
}
