package libkademlia

import (
	"container/list"
	//"fmt"
	//"log"
	//"net"
	//"net/http"
	//"net/rpc"
	//"strconv"
)

// For project 2!
//structure of shortList
type shortList struct {
	visted     map[ID]bool
	pool       *list.List
	closetNode Contact
	result     []Contact
}

//structure of one rpc_find which is used to be processed
type singleResult struct {
	contacts []Contact // rpc find_node call results
	self     Contact   // who sent this result
	err      error
	count    int      //how many rpc_find in this cycle
}

// compare the smaller one
func min(lhs int, rhs int) int {
	if lhs < rhs {
		return lhs
	}
	return rhs
}

//Initialize the shortList
func (sl *shortList) initShortList() {
	sl.visted = make(map[ID]bool)
	sl.pool = list.New()
	sl.closetNode = *(new(Contact))
	sl.result = make([]Contact, 0, 20)
}

//issued one rpc_findnode and insert the result to processchan in order to process in server
func rpc_search(k *Kademlia, c Contact, target ID, processchan chan singleResult, count int) {
	sr := *(new(singleResult))
	result, err := k.DoFindNode(&c, target)
	sr.err = err
	sr.contacts = result
	sr.self = c
	sr.count = count
	processchan <- sr																									//push result to channals
}

//sort shortList.pool, only keep 20 cloest potential node for next cycle
func SortList(ShortList *shortList, target ID) {
	newpool := list.New()
	len := ShortList.pool.Len()
	count := 0
	for i := 0; i < len; i++ {
		closet := ShortList.pool.Front().Value.(Contact)
		for e := ShortList.pool.Front(); e != nil; e = e.Next() {
			if !(target.Xor(closet.NodeID).Less(target.Xor(e.Value.(Contact).NodeID))) {
				closet = e.Value.(Contact)
			}
		}
		newpool.PushBack(closet)
		for e := ShortList.pool.Front(); e != nil; e = e.Next() {
			if e.Value.(Contact).NodeID.Equals(closet.NodeID) {
				ShortList.pool.Remove(e)
			}
		}
		count++
		if count == 20 {																		//only keep 20 nodes
			break
		}
	}
	ShortList.pool = newpool
}

//add the result to the shortList, and add the active process to the result list, and update closetNode
func dealWithSingleResult(myShortList *shortList, sr singleResult, target ID, flag *bool) {
	if sr.err == nil {
		myShortList.result = append(myShortList.result, sr.self)											//add the active process
		for i := 0; i < len(sr.contacts); i++ {
			if _, ok := myShortList.visted[sr.contacts[i].NodeID]; !ok {								//if this node has not been visited
				myShortList.pool.PushFront(sr.contacts[i])
				if !(target.Xor(myShortList.closetNode.NodeID).Less(target.Xor(sr.contacts[i].NodeID))) {	//update closetNode
					myShortList.closetNode = sr.contacts[i]
					*flag = true
				}
			}
		}
	}
}

//a server keep running processing the result of each rpc_find
func (k *Kademlia) start_update_check_service(target ID, myShortList *shortList, processchan chan singleResult,
	poolchan chan []Contact, flagchan chan bool) {
	//process the result of a cycle
	for {
		flag := false
		count := 0
		for {
			para := <-processchan
			dealWithSingleResult(myShortList, para, target, &flag) 	//process the result of single rpc_find
			if len(myShortList.result) == 20 {
				flag = false
				break
			}
			count++
			if count == para.count { 																//if it is the last call of this cycle, break
				break
			}
		}
		//select the next nodes for rpc_find
		if flag {
			next_nodes := make([]Contact, 0, alpha)
			if myShortList.pool.Len() == 0 {
				flag = false
			} else {
				SortList(myShortList, target)														//sort the shortList.pool
				currentPoolSize := myShortList.pool.Len()
				length := alpha
				if currentPoolSize < alpha {
					length = currentPoolSize
				}
				for i := 0; i < length; i++ {
					ele := myShortList.pool.Front()
					next_nodes = append(next_nodes, ele.Value.(Contact))
					myShortList.visted[next_nodes[i].NodeID] = true								//add this node to be visited
					myShortList.pool.Remove(ele)																	//remove from the myShortList.pool
				}
			}
			if len(next_nodes) == 0 {
				flag = false
				break
			}
			go func() {																										//add the next_nodes to the poolchan
				poolchan <- next_nodes
			}()
		}
			go func() {																											//add the result to the resultchan
				flagchan <- flag
			}()
			if flag == false {																							//if meet terminate condition, terminate service
				break
			}
	}
}

// data types for iterative find value
type valShortList struct {
	visted     map[ID]bool
	pool       *list.List
	closetNode Contact
	active     []Contact
	nodesToSto []Contact
	key        ID			// stored key
	val        []byte		// stored value
}

type rpcFindValRes struct {
	self     Contact
	value    []byte
	contacts []Contact
	err      error
}

func (sl *valShortList) initValShortList(kk *Kademlia, key ID) {
	sl.visted = make(map[ID]bool)
	sl.pool = list.New()
	sl.closetNode = kk.SelfContact
	sl.active = make([]Contact, 0, k)
	sl.nodesToSto = make([]Contact, 0, k)
	sl.key = key
	sl.val = nil
}

func getContactsFromPool(reqPoolChan chan bool,
	resPoolChan chan []Contact) (res []Contact) {
	reqPoolChan <- true
	res = <-resPoolChan
	return
}

func dealWithFindValRes(sl *valShortList, res rpcFindValRes) {
	// mark visited node
	// this function only update value or insert contacts into pool  !!
	sl.visted[res.self.NodeID] = true
	if sl.val == nil {
		sl.val = res.value
	}

	if res.err == nil {
		sl.active = append(sl.active, res.self)
	}

	if res.err == nil && res.value == nil {
		// contacted with node but find no value
		sl.nodesToSto = append(sl.nodesToSto, res.self)
		for i := 0; i < len(res.contacts); i++ {
			if _, ok := sl.visted[res.contacts[i].NodeID]; !ok {
				// not contacted before
				sl.pool.PushFront(res.contacts[i])
			}
		}
	}
}

func sortValList(sl *valShortList) {
	newpool := list.New()
	len := sl.pool.Len()
	count := 0
	for i := 0; i < len; i++ {
		closet := sl.pool.Front().Value.(Contact)
		for e := sl.pool.Front(); e != nil; e = e.Next() {
			if !(sl.key.Xor(closet.NodeID).Less(sl.key.Xor(e.Value.(Contact).NodeID))) {
				closet = e.Value.(Contact)
			}
		}
		newpool.PushBack(closet)
		for e := sl.pool.Front(); e != nil; e = e.Next() {
			if e.Value.(Contact).NodeID.Equals(closet.NodeID) {
				sl.pool.Remove(e)
			}
		}
		count++
		if count == 20 {
			break
		}
	}
	sl.pool = newpool
}

func updatePoolAndClosetNode(sl *valShortList) (isChanged bool) {
	// update closet node and sort my nodesToSto and return still running
	// return true if cloest node is changed
	isChanged = false
	if sl.pool.Len() > 0 {
		// sort sl.pool
		sortValList(sl)
		if sl.key.Xor(sl.pool.Front().Value.(Contact).NodeID).Less(sl.key.Xor(sl.closetNode.NodeID)) {
			sl.closetNode = sl.pool.Front().Value.(Contact)
			isChanged = true
		}
	}
	return
}

func valShortListManager(sl *valShortList, reqPoolChan chan bool, resPoolChan chan []Contact,
	isRpcResFinish chan bool, rpcResChan chan rpcFindValRes, mgrCloseChan chan bool) {
	for {
		select {
		// get from shortlist pool
		case <-reqPoolChan:
			length := min(sl.pool.Len(), alpha)
			contacts := make([]Contact, length, length)
			for i := 0; i < length; i++ {
				ele := sl.pool.Front()
				contacts[i] = ele.Value.(Contact)
				sl.pool.Remove(ele)
			}
			resPoolChan <- contacts
		// dealwith rpcfindvalue_result
		case res := <-rpcResChan:
			dealWithFindValRes(sl, res)
			isRpcResFinish <- true
		// terminate this goroutine
		case <-mgrCloseChan:
			return
		}
	}
}

func (kk *Kademlia) rpc_findValue(contact Contact, key ID, rpcFindValResChan chan rpcFindValRes) {
	val, contacts, err := kk.DoFindValue(&contact, key)
	result := rpcFindValRes{contact, val, contacts, err}
	rpcFindValResChan <- result
}
