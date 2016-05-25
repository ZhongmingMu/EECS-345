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
	//"time"
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
	reschan  chan []Contact
	nodeid   ID
	senderid ID
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

type StoVDOType struct {
	vdoID ID
	vdo   VanashingDataObject
}

type GetVDOType struct {
	vdoID   ID
	reschan chan VanashingDataObject
}

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact

	RouteTable []K_Buckets   //RTable, k-buckets
	DataStore  map[ID][]byte //map store the key-value
	VDOStore   map[ID]VanashingDataObject

	RTManagerChan   chan Contact         //channel to store update RTable request
	DataStoreChan   chan KVpair          //channel to store store key-value request
	NodeFindChan    chan FindBucketType  //channel to store find node request
	ContactFindChan chan FindContactType //channel to store find contact request
	ValueFindChan   chan FindValueType   //channel to store find value request

	StoVDOChan chan StoVDOType // channel to StoVDOType
	GetVDOChan chan GetVDOType // channel to getvdotype
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
	// project 3 init vanish data structure
	k.StoVDOChan = make(chan StoVDOType)
	k.GetVDOChan = make(chan GetVDOType)
	k.VDOStore = make(map[ID]VanashingDataObject)
	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}

	////fmt.Printf("%d: begin start server\n", k.SelfContact.Port)

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

	////fmt.Printf("%d: create new kademlia node \n", k.SelfContact.Port)
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
	////fmt.Println(distance.PrefixLen())
	return distance.PrefixLen()
}

// Update  RouteTable(k-buckets)
func (k *Kademlia) UpdateRouteTable(c *Contact) {
	num := FindBucketNum(k.NodeID, c.NodeID) //find the bucket num that possible hold c
	//	//fmt.Printf("%d: insert num: %d \n", k.SelfContact.Port, num)

	l := k.RouteTable[num]

	//	_, erro := k.directFindContact(c.NodeID, num) //find the contact c in this
	//erro := new(Error)

	ll := l.bucket
	erro := false
	for e := ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*Contact).NodeID.Equals(c.NodeID) {
			//fmt.Printf("%d: successful find contact 2: \n", k.SelfContact.Port)
			erro = true
		}
	}

	if erro == true { //if c exist, move it to the tail of bucket
		l.MoveToTail(c)
	} else {
		if !l.CheckFull() { //if c do not exist and the list is not full, add it to the tail
			l.AddTail(c)
			// //fmt.Println(l.bucket.Len())
		} else { //if c do not exist and the list is full
			head := l.GetHead()
			port := head.Port
			host := head.Host

			portstr := strconv.Itoa(int(port))
			client, err := rpc.DialHTTPPath("tcp", host.String()+":"+portstr, rpc.DefaultRPCPath+portstr)
			if err != nil {
				log.Fatal("dialing:", err)
			}

			//_, erro := k.DoPing(head.Host, head.Port) //try to contact the head node
			//	if erro != nil {                          //if head node do not exist, delete it and add the current node to tail
			if client == nil {
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
	////fmt.Printf("%d: starting UpdateHandler \n", k.SelfContact.Port)
	for {
		////fmt.Printf("%d: for \n", k.SelfContact.Port)
		select {
		case c := <-k.RTManagerChan: //handle update k-buckets
			//fmt.Printf("%d: before update RT \n", k.SelfContact.Port)
			k.UpdateRouteTable(&c)
		case p := <-k.DataStoreChan: //handle store date to datamap
			//fmt.Printf("%d: before DataStore \n", k.SelfContact.Port)
			k.UpdateDataStore(&p)
		case f := <-k.NodeFindChan: //handle node finding
			//fmt.Printf("%d: before FindNode \n", k.SelfContact.Port)
			k.findCloestNodes(f.senderid, f.nodeid, f.reschan)
		case cf := <-k.ContactFindChan: //handle contact finding
			//fmt.Printf("%d: before findcontact \n", k.SelfContact.Port)
			k.FindContactHelper(cf.nodeid, cf.reschan)
		case v := <-k.ValueFindChan: //handle value finding
			//fmt.Printf("%d: before findvalue \n", k.SelfContact.Port)
			val, err := k.LocalFindValue(v.searchkey)
			v.reschan <- ValueRes{err, val}
			//	default:
			////fmt.Printf("%d: default \n", k.SelfContact.Port)
			//		continue
		case stoVdo := <-k.StoVDOChan:
			k.VDOStore[stoVdo.vdoID] = stoVdo.vdo

		case getVDOReq := <-k.GetVDOChan:
			vdo, _ := k.VDOStore[getVDOReq.vdoID]
			getVDOReq.reschan <- vdo

		}
	}
}

//handle finding nodes func---find k cloest nodes
func (kk *Kademlia) findCloestNodes(senderid ID, nodeid ID, reschan chan []Contact) {
	nodes := ([]Contact{}) //result list

	closestnum := FindBucketNum(kk.NodeID, nodeid) //find bucket num storing nodeid
	count := 0
	diff := 1

	//fmt.Printf("%d: closetnumber  %d\n", kk.SelfContact.Port, closestnum)

	//add this bucket's nodes to the result list
	for e := kk.RouteTable[closestnum].bucket.Front(); e != nil; e = e.Next() {
		//fmt.Printf("%d: here count %d\n", kk.SelfContact.Port, count)
		//fmt.Printf("%d: list length %d\n", kk.SelfContact.Port, kk.RouteTable[closestnum].bucket.Len())
		//fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, FormatTrans(e.Value.(*Contact)).Port)
		////fmt.Printf("%d: port  %d\n", kk.SelfContact.Port, nodes[0].Port)
		if !FormatTrans(e.Value.(*Contact)).NodeID.Equals(senderid) {
			nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
			count = count + 1
		}
	}

	//if the closet bucket has < k nodes, find neigh nodes until find k nodes or find all the buckets

	for count < k {
		if closestnum+diff < b {
			for e := kk.RouteTable[closestnum+diff].bucket.Front(); e != nil; e = e.Next() {
				if !FormatTrans(e.Value.(*Contact)).NodeID.Equals(senderid) {
					nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
					count = count + 1
					if count >= k-1 {
						break
					}
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
				//fmt.Printf("%d: here count %d \n", kk.SelfContact.Port, count)
				if !FormatTrans(e.Value.(*Contact)).NodeID.Equals(senderid) {
					nodes = append(nodes, FormatTrans(e.Value.(*Contact)))
					count = count + 1
					if count >= k-1 {
						break
					}
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
	// //fmt.Printf("%d: before push to findchan \n", k.SelfContact.Port)
	k.ContactFindChan <- contactfind //put the finding contact request to channel to find thread safely
	// //fmt.Printf("%d: after push to findchan \n", k.SelfContact.Port)
	res := <-contactfind.reschan //extract back the results
	// //fmt.Printf("%d:after get from findchan \n", k.SelfContact.Port)
	return res.contact, res.err
}

// Direct find contact without handler
func (k *Kademlia) directFindContact(nodeId ID) (*Contact, error) {
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	} else {
		num := FindBucketNum(k.NodeID, nodeId)
		//fmt.Printf("%d: find num: %d\n", k.SelfContact.Port, num)
		l := k.RouteTable[num].bucket
		// //fmt.Println(l.Len())
		for e := l.Front(); e != nil; e = e.Next() {
			if e.Value.(*Contact).NodeID.Equals(nodeId) {
				//fmt.Printf("%d: successful find contact 2: \n", k.SelfContact.Port)
				return e.Value.(*Contact), nil
			}
		}
	}
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}

// handle contact finding request func
func (k *Kademlia) FindContactHelper(nodeId ID, reschan chan ContactErr) {
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
	//fmt.Printf("%d: finish dialhttpath: \n", k.SelfContact.Port)
	//client, err := rpc.DialHTTP("tcp", host.String()+":"+strconv.FormatInt(int64(port), 10))
	if err != nil {
		//fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return nil, &CommandFailed{
			"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
	}

	err = client.Call("KademliaRPC.Ping", ping, &pong) //RPC Ping function

	defer func() {
		client.Close()
	}()

	//fmt.Printf("%d: finish remote ping: \n", k.SelfContact.Port)
	if err == nil {
		//fmt.Printf("%d: call succesfull: \n", k.SelfContact.Port)
		//fmt.Printf("%d: pong sender port: %d \n", k.SelfContact.Port, pong.Sender.Port)

		k.RTManagerChan <- pong.Sender //set update k-buckets table request to update request contact
		//c1 := <- k.RTManagerChan
		//		//fmt.Printf("%d: channel %d \n", k.SelfContact.Port, c1.Port)
		//fmt.Printf("%d: push to channel \n", k.SelfContact.Port)
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
		//fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
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
		//fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
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
	for _, result := range res.Nodes {
		k.RTManagerChan <- result
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
		//fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
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
		return nil, nil, &CommandFailed{"message ID not match"}
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

// for project 2, the help function is in iterative_lib.go
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	myShortList := *(new(shortList))
	poolchan := make(chan []Contact)
	flagchan := make(chan bool)
	processchan := make(chan singleResult)
	reschan := make(chan []Contact)
	myShortList.initShortList() //initial the shortList
	//local find ( first find)
	fbt := FindBucketType{reschan, id, k.SelfContact.NodeID}
	go func() {
		k.NodeFindChan <- fbt
	}()
	contacts := <-reschan
	if len(contacts) < 1 {
		return nil, &CommandFailed{"No Contact is Found"}
	}
	//find the closetNode of the first local find nodes
	myShortList.closetNode = contacts[0]
	for i := 0; i < len(contacts); i++ {
		if !(id.Xor(myShortList.closetNode.NodeID).Less(id.Xor(contacts[i].NodeID))) {
			myShortList.closetNode = contacts[i]
		}
	}
	// add initial search nodes to pool channals, and mark them visited
	firstpoll := make([]Contact, 0, 20)
	firstpoll = append(firstpoll, myShortList.closetNode)
	myShortList.visted[myShortList.closetNode.NodeID] = true
	count := 1
	for i := 0; i < len(contacts); i++ {
		if _, ok := myShortList.visted[contacts[i].NodeID]; !ok {
			firstpoll = append(firstpoll, contacts[i])
			myShortList.visted[contacts[i].NodeID] = true
			count++
		}
		if count == alpha {
			break
		}
	}
	go func() {
		poolchan <- firstpoll
	}()

	//start start_update_check_service
	go k.start_update_check_service(id, &myShortList, processchan, poolchan, flagchan)
	//start the iterative find
	for {
		count := 0
		flag := true
		next_nodes := <-poolchan //get the nodes from where to find
		for i := 0; i < len(next_nodes); i++ {
			go rpc_search(k, next_nodes[i], id, processchan, len(next_nodes))
		}
		count++
		flag = <-flagchan //extract the result of one cycle
		if flag == false {
			break
		}
	}
	//fill the shortlist if possible
	if len(myShortList.result) < 20 && myShortList.pool.Len() > 0 {
		for e := myShortList.pool.Front(); e != nil; e = e.Next() {
			c := e.Value.(Contact)
			if _, ok := myShortList.visted[c.NodeID]; !ok { //if node is active and has not been visted, add them to result
				_, err := k.DoFindNode(&c, id)
				if err == nil {
					myShortList.result = append(myShortList.result, c)
					myShortList.visted[c.NodeID] = true
				}
				if len(myShortList.result) == 20 {
					break
				}
			}
		}
	}
	if len(myShortList.result) == 0 {
		return nil, &CommandFailed{"No Contact is Found"}
	}

	return myShortList.result, nil
}

func (kk *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	rcvdContacts := make([]Contact, 0, k)

	triples, err := kk.DoIterativeFindNode(key)
	if err != nil {
		return nil, &CommandFailed{"No Value is Stored"}
	}

	for i := 0; i < len(triples); i++ {
		err := kk.DoStore(&triples[i], key, value)
		if err == nil {
			rcvdContacts = append(rcvdContacts, triples[i])
		}
	}

	if len(rcvdContacts) == 0 {
		return nil, &CommandFailed{"No Value is Stored"}
	}
	return rcvdContacts, nil
}

func (kk *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	myShortList := new(valShortList)
	myShortList.initValShortList(kk, key)
	value, err = kk.LocalFindValue(key)
	if err == nil {
		// already find value in local store
		return
	}

	localreschan := make(chan []Contact)
	fbt := FindBucketType{localreschan, key, kk.SelfContact.NodeID}
	// go func() {
	// 	k.NodeFindChan <- fbt
	// }()
	kk.NodeFindChan <- fbt
	contacts := <-localreschan

	// valshortlist manager
	var (
		// init my channel
		rpcFindValResChan = make(chan rpcFindValRes)
		mgrCloseChan      = make(chan bool)
		reqPoolChan       = make(chan bool)
		resPoolChan       = make(chan []Contact)
		isChangedChan     = make(chan bool)
	)

	go valShortListManager(myShortList, reqPoolChan,
		resPoolChan, isChangedChan,
		rpcFindValResChan, mgrCloseChan)
	defer func() {
		// close the valshortlistmanager when close this funciton return
		mgrCloseChan <- true
	}()

	cs := contacts[0:min(len(contacts), alpha)]

	for running := true; running; cs = getContactsFromPool(reqPoolChan, resPoolChan) {
		for i := 0; i < len(cs); i++ {
			// launch rpcFindVal routines
			//fmt.Println(i)
			go kk.rpc_findValue(cs[i], key, rpcFindValResChan)
		}

		// running = false
		for i := 0; i < len(cs); i++ {
			// wait for len(cs) goroutine return
			<-isChangedChan
			// synchronize here wait for 3 rpc find value finish
		}

		running = updatePoolAndClosetNode(myShortList)
		// running = false

		if myShortList.val == nil && running == false && len(myShortList.active) < 20 && myShortList.pool.Len() > 0 {
			running = true
		}

	}
	myShortList.active = myShortList.active[:min(len(myShortList.active), 20)]
	//fmt.Println(len(myShortList.active))

	if myShortList.val != nil {
		// store value in closest contact
		length := len(myShortList.nodesToSto)
		if length > 0 {
			nodeToSto := myShortList.nodesToSto[0]
			for i := 1; i < length; i++ {
				if key.Xor(myShortList.nodesToSto[i].NodeID).Less(key.Xor(nodeToSto.NodeID)) {
					nodeToSto = myShortList.nodesToSto[i]
				}
			}
			kk.DoStore(&nodeToSto, key, myShortList.val)
		}

		return myShortList.val, nil
	}
	return nil, &CommandFailed{"value not found"}
}

// For project 3!
func (k *Kademlia) Vanish(vdoID ID, data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	vdo := k.VanishData(data, numberKeys, threshold, timeoutSeconds)
	k.StoreVDO(vdoID, vdo)
	return
}

func (k *Kademlia) Unvanish(nodeID ID, vdoID ID) (data []byte) {
	req := GetVDORequest{k.SelfContact, vdoID, NewRandomID()}
	res := new(GetVDOResult) //set the Result
	contact, err := k.directFindContact(nodeID)
	if err != nil {
		fmt.Printf("Can not find nodeID")
		return nil
	}
	portStr := strconv.Itoa(int(contact.Port))
	firstPeerStr := contact.Host.String() + ":" + portStr

	client, err := rpc.DialHTTPPath("tcp", firstPeerStr, rpc.DefaultRPCPath+portStr) //set the connection

	if err != nil {
		//fmt.Printf("%d: connection failed: \n", k.SelfContact.Port)
		//log.Fatal("dialing:", err)
		return nil, &CommandFailed{
			"Unable to FindNode " + fmt.Sprintf("%s:%v", contact.Host.String(), contact.Port)}
	}

	err = client.Call("KademliaRPC.GetVDO", req, &res) //RPC FindNode func
	defer func() {
		client.Close()
	}()

	data := k.UnvanishData(res.VDO)
	return data
}
