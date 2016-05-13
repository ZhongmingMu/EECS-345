package libkademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"container/list"
	"fmt"
	"log"
	"math"
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

//key-value struct
type KVpair struct {
	key   ID
	value []byte
}

//findNode request struct
type FindBucketType struct {
	reschan chan []Contact
	nodeid  ID
}

//Contactfind Result struct
type ContactErr struct {
	contact *Contact
	err     error
}

//ContactFind request struct
type FindContactType struct {
	reschan chan ContactErr
	nodeid  ID
}

//Valuefind result struct
type ValueRes struct {
	err   error
	value []byte
}

//ValueFind request struct
type FindValueType struct {
	reschan   chan ValueRes
	searchkey ID
}

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact

	RouteTable []K_Buckets   //RTable, k-buckets
	DataStore  map[ID][]byte //map store the key-value

	RTManagerChan   chan Contact         //channel to store update RTable request
	DataStoreChan   chan KVpair          //channel to store store key-value request
	NodeFindChan    chan FindBucketType  //channel to store find node request
	ContactFindChan chan FindContactType //channel to store find contact request
	ValueFindChan   chan FindValueType   //channel to store find value request
}

func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID

	// TODO: Initialize other state here as you add functionality.

	//initial k-buckets
	k.RouteTable = make([]K_Buckets, b)
	for i := 0; i < b; i++ {
		k.RouteTable[i] = *NewKBuckets(20, i)
	}
	//initial channels
	k.RTManagerChan = make(chan Contact)
	k.DataStoreChan = make(chan KVpair)
	k.DataStore = make(map[ID][]byte)
	k.NodeFindChan = make(chan FindBucketType)
	k.ContactFindChan = make(chan FindContactType)
	k.ValueFindChan = make(chan FindValueType)
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}

	fmt.Printf("%d: begin start server\n", k.SelfContact.Port)

	s.HandleHTTP(rpc.DefaultRPCPath+port, rpc.DefaultDebugPath+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// start handler
	go k.UpdateHandler()

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

	fmt.Printf("%d: create new kademlia node \n", k.SelfContact.Port)
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

//find the bucket num storing rhs
func FindBucketNum(lhs ID, rhs ID) int {
	distance := lhs.Xor(rhs)
	//fmt.Println(distance.PrefixLen())
	return distance.PrefixLen()
}

// Update  RouteTable(k-buckets)
func (k *Kademlia) UpdateRouteTable(c *Contact) {
	num := FindBucketNum(k.NodeID, c.NodeID) //find the bucket num that possible hold c
	fmt.Printf("%d: insert num: %d \n", k.SelfContact.Port, num)

	l := k.RouteTable[num]

	_, erro := k.directFindContact(c.NodeID) //find the contact c in this

	if erro == nil { //if c exist, move it to the tail of bucket
		l.MoveToTail(c)
	} else {
		if !l.CheckFull() { //if c do not exist and the list is not full, add it to the tail
			l.AddTail(c)
			// fmt.Println(l.bucket.Len())
		} else { //if c do not exist and the list is full
			head := l.GetHead()
			_, erro := k.DoPing(head.Host, head.Port) //try to contact the head node
			if erro != nil {                          //if head node do not exist, delete it and add the current node to tail
				l.RemoveHead()
				l.AddTail(c)
			}
		}
	}
}

//update the datamap
func (k *Kademlia) UpdateDataStore(p *KVpair) {
	k.DataStore[p.key] = p.value
}

// Handler aimed to make the whole program thread safe
func (k *Kademlia) UpdateHandler() {
	//fmt.Printf("%d: starting UpdateHandler \n", k.SelfContact.Port)
	for {
		//fmt.Printf("%d: for \n", k.SelfContact.Port)
		select {
		case c := <-k.RTManagerChan: //handle update k-buckets
			fmt.Printf("%d: before update RT \n", k.SelfContact.Port)
			k.UpdateRouteTable(&c)
		case p := <-k.DataStoreChan: //handle store date to datamap
			fmt.Printf("%d: before DataStore \n", k.SelfContact.Port)
			k.UpdateDataStore(&p)
		case f := <-k.NodeFindChan: //handle node finding
			fmt.Printf("%d: before FindNode \n", k.SelfContact.Port)
			k.findCloestNodes(f.nodeid, f.reschan)
		case cf := <-k.ContactFindChan: //handle contact finding
			fmt.Printf("%d: before findcontact \n", k.SelfContact.Port)
			k.FindContactHelper(cf.nodeid, cf.reschan)
		case v := <-k.ValueFindChan: //handle value finding
			fmt.Printf("%d: before findvalue \n", k.SelfContact.Port)
			val, err := k.LocalFindValue(v.searchkey)
			v.reschan <- ValueRes{err, val}
		default:
			//fmt.Printf("%d: default \n", k.SelfContact.Port)
			continue
		}
	}
}

//handle finding nodes func---find k cloest nodes
func (kk *Kademlia) findCloestNodes(nodeid ID, reschan chan []Contact) {
	nodes := ([]Contact{}) //result list

	closestnum := FindBucketNum(kk.NodeID, nodeid) //find bucket num storing nodeid
	count := 0
	diff := 1

	fmt.Printf("%d: closetnumber  %d\n", kk.SelfContact.Port, closestnum)

	//add this bucket's nodes to the result list
	for e := kk.RouteTable[closestnum].bucket.Front(); e != nil; e = e.Next() {
		fmt.Printf("%d: here count %d\n", kk.SelfContact.Port, count)
		fmt.Printf("%d: list length %d\n", kk.SelfContact.Port, kk.RouteTable[closestnum].bucket.Len())
		fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, FormatTrans(e.Value.(*Contact)).Port)
		//fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, nodes[0].Port)
		nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
		count = count + 1
	}

	//if the closet bucket has < k nodes, find neigh nodes until find k nodes or find all the buckets

	for count < k {
		if closestnum+diff < b {
			for e := kk.RouteTable[closestnum+diff].bucket.Front(); e != nil; e = e.Next() {
				nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
				count = count + 1
				if count >= k-1 {
					break
				}
			}
		}
		if closestnum+diff >= b {
			break
		}
		diff = diff + 1
	}

	diff = 1
	for count < k {
		//find closet - 1 bucket
		if closestnum-diff >= 0 {
			for e := kk.RouteTable[closestnum-diff].bucket.Front(); e != nil; e = e.Next() {
				fmt.Printf("%d: here count %d \n", kk.SelfContact.Port, count)
				nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
				count = count + 1
				if count >= k-1 {
					break
				}
			}
		}

		if closestnum-diff < 0 {
			break
		}
		diff = diff + 1
	}

	reschan <- nodes //put the result back to res channel
}

//find contact with nodeID ID
func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	reschan := make(chan ContactErr) //setup a temp channel to store the result
	contactfind := FindContactType{reschan, nodeId}
	// fmt.Printf("%d: before push to findchan \n", k.SelfContact.Port)
	k.ContactFindChan <- contactfind //put the finding contact request to channel to find thread safely
	// fmt.Printf("%d: after push to findchan \n", k.SelfContact.Port)
	res := <-contactfind.reschan //extract back the results
	// fmt.Printf("%d:after get from findchan \n", k.SelfContact.Port)
	return res.contact, res.err
}

// Direct find contact without handler
func (k *Kademlia) directFindContact(nodeId ID) (*Contact, error) {
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	} else {
		num := FindBucketNum(k.NodeID, nodeId)
		fmt.Printf("%d: find num: %d\n", k.SelfContact.Port, num)
		l := k.RouteTable[num].bucket
		// fmt.Println(l.Len())
		for e := l.Front(); e != nil; e = e.Next() {
			if e.Value.(*Contact).NodeID.Equals(nodeId) {
				fmt.Printf("%d: successful find contact 2: \n", k.SelfContact.Port)
				return e.Value.(*Contact), nil
			}
		}
	}
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}

// handle contact finding request func
func (k *Kademlia) FindContactHelper(nodeId ID, reschan chan ContactErr) {
	/*
		if nodeId == k.SelfContact.NodeID { 												//if finding itself, directly return
			reschan <- ContactErr{&k.SelfContact, nil}
			return
		} else {
			num := FindBucketNum(k.NodeID, nodeId)											//find the closet bucket of the comtact
			fmt.Printf("%d: find num: %d\n", k.SelfContact.Port, num)
			l := k.RouteTable[num].bucket
			//fmt.Println(l.Len())
			for e := l.Front(); e != nil; e = e.Next() { 									//traverse the buckets to find the contact
				if e.Value.(* Contact).NodeID.Equals(nodeId) {
					fmt.Printf("%d: successful find contact 2: \n", k.SelfContact.Port)
					reschan <- ContactErr{e.Value.(* Contact), nil}
					return
				}
			}
		}
		reschan <- ContactErr{nil, &ContactNotFoundError{nodeId, "Not found"}} 			  	//do not find the contact, return ERR
	*/
	c, err := k.directFindContact(nodeId)
	reschan <- ContactErr{c, err}
}

type CommandFailed struct {
	msg string
}

func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}

func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement
	ping := PingMessage{k.SelfContact, NewRandomID()} //setup PING msg
	pong := new(PongMessage)                          //setup PONG msg

	//client, err := rpc.DialHTTPPath("tcp", host.String() + ":" + strconv.Itoa(int (port)),
	//	                 rpc.DefaultRPCPath + strconv.Itoa(int (port)))
	firstPeerStr := host.String() + ":" + strconv.Itoa(int(port))
	//client, err := rpc.DialHTTP("tcp", firstPeerStr)
	client, err := rpc.DialHTTPPath("tcp", firstPeerStr, rpc.DefaultRPCPath+strconv.Itoa(int(port))) // Set connection
	fmt.Printf("%d: finish dialhttpath: \n", k.SelfContact.Port)
	//client, err := rpc.DialHTTP("tcp", host.String()+":"+strconv.FormatInt(int64(port), 10))
	if err != nil {
		fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return nil, &CommandFailed{
			"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}

	err = client.Call("KademliaRPC.Ping", ping, &pong) //RPC Ping function
	defer func() {
		client.Close()
	}()

	fmt.Printf("%d: finish remote ping: \n", k.SelfContact.Port)
	if err == nil {
		fmt.Printf("%d: call succesfull: \n", k.SelfContact.Port)
		fmt.Printf("%d: pong sender port: %d \n", k.SelfContact.Port, pong.Sender.Port)

		k.RTManagerChan <- pong.Sender //set update k-buckets table request to update request contact
		//c1 := <- k.RTManagerChan
		//		fmt.Printf("%d: channel %d \n", k.SelfContact.Port, c1.Port)
		fmt.Printf("%d: push to channel \n", k.SelfContact.Port)
		return &(pong.Sender), nil
	} else {
		return nil, &CommandFailed{
			"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}

}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	// TODO: Implement
	storeReq := StoreRequest{k.SelfContact, NewRandomID(), key, value} //Set Store Request Msg
	storeRes := new(StoreResult)                                       //setup Result

	portStr := strconv.Itoa(int(contact.Port))
	firstPeerStr := contact.Host.String() + ":" + portStr

	client, err := rpc.DialHTTPPath("tcp", firstPeerStr, rpc.DefaultRPCPath+portStr) //set the connection

	if err != nil {
		fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return &CommandFailed{
			"Unable to ping " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}
	err = client.Call("KademliaRPC.Store", storeReq, &storeRes) //RPC Store method
	defer func() {
		client.Close()
	}()

	if storeRes.Err != nil {
		return storeRes.Err
	}

	if storeReq.MsgID.Equals(storeRes.MsgID) { //Update the receiver contact
		k.RTManagerChan <- *contact
	}
	return nil
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	req := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey} //set the request Msg
	res := new(FindNodeResult)                                      //set the Result

	portStr := strconv.Itoa(int(contact.Port))
	firstPeerStr := contact.Host.String() + ":" + portStr

	client, err := rpc.DialHTTPPath("tcp", firstPeerStr, rpc.DefaultRPCPath+portStr) //set the connection

	if err != nil {
		fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return nil, &CommandFailed{
			"Unable to FindNode " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}

	err = client.Call("KademliaRPC.FindNode", req, &res) //RPC FindNode func
	defer func() {
		client.Close()
	}()

	if res.Err != nil {
		return nil, res.Err
	}

	if req.MsgID.Equals(res.MsgID) { //set update k-buckets table request to update request contact
		k.RTManagerChan <- *contact
	}
	return res.Nodes, nil
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	req := FindValueRequest{k.SelfContact, NewRandomID(), searchKey} //set the request Msg
	res := new(FindValueResult)                                      //set the Result

	portStr := strconv.Itoa(int(contact.Port))
	firstPeerStr := contact.Host.String() + ":" + portStr

	client, err := rpc.DialHTTPPath("tcp", firstPeerStr, rpc.DefaultRPCPath+portStr) //set the connection

	if err != nil {
		fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return nil, nil, &CommandFailed{
			"Unable to FindValue " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}

	err = client.Call("KademliaRPC.FindValue", req, &res) //RPC FindValue func
	defer func() {
		client.Close()
	}()

	k.RTManagerChan <- *contact //set update k-buckets table request to update request contact
	if !res.MsgID.Equals(req.MsgID) {
		return nil, nil, &CommandFailed{"Not implemented"}
	}

	value = res.Value
	contacts = res.Nodes
	err = res.Err

	return
}

//Local Find value in datamap
func (k *Kademlia) LocalFindValue(searchKey ID) ([]byte, error) {
	// TODO: Implement

	if val, ok := k.DataStore[searchKey]; ok {
		return val, nil
	} else {
		return []byte(""), &CommandFailed{"Value not exists"}
	}

}

type shortList struct {
	visted     map[ID]bool
	pool       *list.List
	closetNode Contact
	result     []Contact
}

type singleResult struct {
	contacts  []Contact // rpc find_node call results
	self      Contact   // who sent this result
	err       error
	iden_last bool
}

//because each cycle create a new shortlist, should return it with the flag
type returnType struct {
	flag       bool
	resultList shortList
}

func (sl *shortList) initShortList(k *Kademlia) {
	sl.visted = make(map[ID]bool)
	sl.pool = list.New()
	sl.closetNode = *(new(Contact))
	sl.result = make([]Contact, 0, 20)

}

func rpc_search(k *Kademlia, c Contact, target ID, processchan chan singleResult, last_identi bool) {
	sr := *(new(singleResult))
	result, err := k.DoFindNode(&c, target)
	sr.err = err
	sr.contacts = result
	sr.self = c
	sr.iden_last = last_identi

	processchan <- sr
}

func dealWithSingleResult(ShortList *shortList, sr singleResult, target ID, flag *bool) shortList {
	//myShortList.visted[sr.self.NodeID] = true
	myShortList := *ShortList
	if sr.err == nil {
		myShortList.result = append(myShortList.result, sr.self)
		for i := 0; i < len(sr.contacts); i++ {
			if _, ok := myShortList.visted[sr.contacts[i].NodeID]; !ok {
				myShortList.pool.PushFront(sr.contacts[i])
				if !(target.Xor(myShortList.closetNode.NodeID).Less(target.Xor(sr.contacts[i].NodeID))) {
					myShortList.closetNode = sr.contacts[i]
					*flag = true
				}
			}
		}
		//myShortList.result = append(myShortList.result, sr.self)
	}
	return myShortList
}

func (k *Kademlia) start_update_check_service(target ID, myShortList shortList, processchan chan singleResult,
	poolchan chan []Contact, flagchan chan returnType) {

	for {
		flag := false
		for {
			para := <-processchan
			myShortList = dealWithSingleResult(&myShortList, para, target, &flag) //it's not the original myShortList
			if len(myShortList.result) == 20 {
				flag = false
				break
			}
			if para.iden_last == true { //if it is the last call of this cycle, break
				break
			}
		}
		if flag {
			next_nodes := make([]Contact, 0, alpha)
			if myShortList.pool.Len() == 0 {
				flag = false
			} else {
				currentPoolSize := myShortList.pool.Len()
				length := alpha
				if currentPoolSize < alpha {
					length = currentPoolSize
				}
				for i := 0; i < length; i++ {
					ele := myShortList.pool.Front()
					next_nodes = append(next_nodes, ele.Value.(Contact))
					myShortList.visted[next_nodes[i].NodeID] = true
					myShortList.pool.Remove(ele)
				}
			}
			if len(next_nodes) == 0 {
				flag = false
				break
			}
			go func() {
				poolchan <- next_nodes
			}()
		}
		singleCycleResult := returnType{flag, myShortList} //return the flag and the new myShortList in case it is the final result
		go func() {
			flagchan <- singleCycleResult
		}()
		if flag == false {
			// we should stop this go routine
			break
		}
	}
}

// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	// init shortlist
	myShortList := *(new(shortList))
	//channals
	poolchan := make(chan []Contact)
	flagchan := make(chan returnType)
	processchan := make(chan singleResult)

	myShortList.initShortList(k)
	//	myShortList.result[9]
	// find local node
	reschan := make(chan []Contact)
	fbt := FindBucketType{reschan, id}
	go func() {
		k.NodeFindChan <- fbt
	}()
	contacts := <-reschan
	if len(contacts) < 1 {
		return nil, &CommandFailed{"node not found\n"}
	}

	myShortList.closetNode = contacts[0]
	for i := 0; i < len(contacts); i++ {
		myShortList.pool.PushFront(contacts[i])
		if !(id.Xor(myShortList.closetNode.NodeID).Less(id.Xor(contacts[i].NodeID))) {
			myShortList.closetNode = contacts[i]
		}
	}
	// add initial search nodes to pool channals
	firstpoll := make([]Contact, 0, 20)
	len1 := int(math.Min(float64(alpha), float64(len(contacts))))
	for i := 0; i < len1; i++ {
		firstpoll = append(firstpoll, contacts[i])
	}
	go func() {
		poolchan <- firstpoll
	}()
	//start start_update_check_service
	go k.start_update_check_service(id, myShortList, processchan, poolchan, flagchan)

	for {
		flag := true
		//poll next alpha nodes
		next_nodes := <-poolchan
		//alpha go routine
		for i := 0; i < len(next_nodes); i++ {
			if i == len(next_nodes)-1 {
				go rpc_search(k, next_nodes[i], id, processchan, true)
			} else {
				go rpc_search(k, next_nodes[i], id, processchan, false)
			}
		}
		//get result judge return
		// update flag
		result := <-flagchan
		flag = result.flag //get the flag
		if flag == false {
			myShortList.result = result.resultList.result // get the final result
			break
		}
	}
	fmt.Print(len(myShortList.result))
	// results1 := [len(shortList.result)]Contact
	// copy(results1[:], shortList.result[:])
	//	shortList.result[0]

	return myShortList.result, nil
	// return nil, &CommandFailed{"Not implemented"}
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	contacts, err := k.DoIterativeFindNode(key)
	if err != nil {
		return nil, &CommandFailed{"Node not found"}
	}
	results := make([]Contact, 0, 20)

	j := 0
	for i := 0; i < len(contacts); i++ {
		erro := k.DoStore(&k.SelfContact, contacts[i].NodeID, value)
		if erro == nil {
			results[j] = contacts[i]
			j++
		}
	}
	return results, nil
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
