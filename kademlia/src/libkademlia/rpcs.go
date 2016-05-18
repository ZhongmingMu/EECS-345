package libkademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	//"fmt"
	"net"
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

//local PING call Function
func (k *KademliaRPC) Ping(ping PingMessage, pong *PongMessage) error {

	//fmt.Println(k.kademlia.SelfContact.Port)
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID) //set Pong Msg ID
	// Specify the sender
	pong.Sender = k.kademlia.SelfContact //set Pong Msg Sender
	// Update contact, etc
	k.kademlia.RTManagerChan <- ping.Sender //update sender in RTtable(k-bucket)
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

	k.kademlia.DataStoreChan <- KVpair{req.Key, req.Value} //Thread calling store function storing pair
	k.kademlia.RTManagerChan <- req.Sender                 //update sender in RTtable(k-bucket)

	res.MsgID = CopyID(req.MsgID) //set res's msg ID
	res.Err = nil                 //set res's msg Error

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

//transfer format of contact
func FormatTrans(c *Contact) Contact {
	newcontact := new(Contact)
	newcontact.Host = c.Host
	newcontact.NodeID = c.NodeID
	newcontact.Port = c.Port

	return *newcontact

}

func (k *KademliaRPC) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	nodeid := req.NodeID            //extract nodeid to be found
	reschan := make(chan []Contact) //setup a channel to receive the result
	res.MsgID = CopyID(req.MsgID)   //set res's ID

	//res.Nodes = k.kademlia.findCloestNodes(nodeid, nodes)
	//put in the findbucket channel in order to find the nodes
	nodetype := FindBucketType{reschan, nodeid, req.Sender.NodeID}
	k.kademlia.NodeFindChan <- nodetype
	// fmt.Printf("%d: ??????? ", k.kademlia.SelfContact.Port)
	res.Nodes = <-nodetype.reschan //extract the result
	// fmt.Printf("%d: !!!!!!! ", k.kademlia.SelfContact.Port)
	res.Err = nil

	k.kademlia.RTManagerChan <- req.Sender //update the k-buckets
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
	searchid := req.Key //extract the key id

	res.MsgID = CopyID(req.MsgID)  //set the res's ID
	reschan := make(chan ValueRes) //setup a temp channel to receive result
	fv := FindValueType{reschan, searchid}

	k.kademlia.ValueFindChan <- fv //put request to channel to find value thread safely

	vr := <-reschan    //extract the result
	if vr.err == nil { //if the value is found,return
		res.Value = vr.value
		res.Err = nil
		res.Nodes = nil
	} else { // if the value is not found, return k closet buckets
		rc := make(chan []Contact)
		nodetype := FindBucketType{rc, searchid, req.Sender.NodeID}
		k.kademlia.NodeFindChan <- nodetype //findNOde()
		res.Nodes = <-nodetype.reschan
		res.Value = nil
		res.Err = nil
	}

	k.kademlia.RTManagerChan <- req.Sender //update the k-buckets

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
