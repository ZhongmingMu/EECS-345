package libkademlia

import (
	//"bytes"
	"fmt"
	//"net"
	"strconv"
	"testing"
//	"time"
)

// func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
// 	hostString, portString, err := net.SplitHostPort(laddr)
// 	if err != nil {
// 		return
// 	}
// 	ipStr, err := net.LookupHost(hostString)
// 	if err != nil {
// 		return
// 	}
// 	for i := 0; i < len(ipStr); i++ {
// 		ip = net.ParseIP(ipStr[i])
// 		if ip.To4() != nil {
// 			break
// 		}
// 	}
// 	portInt, err := strconv.Atoi(portString)
// 	port = uint16(portInt)
// 	return
// }



// EXTRACREDIT
// Check out the correctness of DoIterativeFindNode function
func TestIterativeFindNode1(t *testing.T) {
tree_node := make([]*Kademlia, 30)
address := make([]string, 30)
for i := 0; i < 30; i++ {
	address[i] = "localhost:" + strconv.Itoa(7696+i)
	tree_node[i] = NewKademlia(address[i])
}

//30 nodes ping each other
for i := 0; i < 30 ; i++ {
	for j := 0; j < 30; j++ {
		host_number, port_number, _ := StringToIpPort(address[j])
		tree_node[i].DoPing(host_number, port_number)
	}
}

//find node[19], start from node 0
contacts, _ := tree_node[0].DoIterativeFindNode(tree_node[19].SelfContact.NodeID)
count := 0
//check the result
for i := 0; i < len(contacts); i++ {
	if(contacts[i].NodeID.Equals(tree_node[19].SelfContact.NodeID)) {
		count ++
	}
	fmt.Print(contacts[i].NodeID)
}
if(count != 1) {
	t.Error("the result is not true")
}
}

// Check out the correctness of DoIterativeFindNode function
func TestIterativeFindNode2(t *testing.T) {
	total_num := 40											//total num mush >= 20
	tree_node := make([]*Kademlia, total_num)
	instance1 := NewKademlia("localhost:7599")										//starting node
	host1, port1, _ := StringToIpPort("localhost:7599")

	findId := total_num - 20
	//initialize the nodes
	for i := 0; i < total_num; i++ {
		address := "localhost:" + strconv.Itoa(7600+i)
		tree_node[i] = NewKademlia(address)
		tree_node[i].DoPing(host1, port1)														//every node ping instance1
	}
	target := tree_node[findId].SelfContact.NodeID								//id to be found

	//node less that findId ping the findId, not all ping the findId node
	for i := 0; i < findId; i++ {
			tree_node[findId].DoPing(tree_node[i].SelfContact.Host, tree_node[i].SelfContact.Port)
	}

	result, err := instance1.DoIterativeFindNode(target)						//start from instance1,find target
	if err != nil {
		t.Error(err.Error())
	}

	if result == nil || len(result) == 0 {
		t.Error("No contacts were found")
	}

	//check the result
	count := 0
	for _, value := range result {
		if value.NodeID.Equals(target) {
			count++
		}
	}

	//if only Contains one same node, test pass
	if count != 1 {
		t.Error("test failed")
	}
}

// EXTRACREDIT
//Check out the Correctness of DoIterativeStore
func TestIterativeStore(t *testing.T) {
	// tree structure;
	// A->B->tree->tree2
	/*
	          C
	      /
	   A-B -- D
	       \
	          E
	*/
	instance1 := NewKademlia("localhost:7506")
	instance2 := NewKademlia("localhost:7507")
	host2, port2, _ := StringToIpPort("localhost:7507")
	instance1.DoPing(host2, port2)

	//Build the  A->B->Tree structure
	tree_node := make([]*Kademlia, 20)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(7508+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}
	//implement DoIterativeStore, and get the the result
	value := []byte("Hello world")
	key := NewRandomID()
	contacts, err := instance1.DoIterativeStore(key, value)
	//the number of contacts store the value should be 20
	if err != nil || len(contacts) != 20 {
		t.Error("Error doing DoIterativeStore")
	}
	//Check all the 22 nodes,
	//find out the number of nodes that contains the value
	count := 0
	// check tree_nodes[0~19]
	for i := 0; i < 20; i++ {
		result, err := tree_node[i].LocalFindValue(key)
		if result != nil && err == nil {
			count++
		}
	}
	//check instance2
	result, err := instance2.LocalFindValue(key)
	if result != nil && err == nil {
		count++
	}
	//check instance1
	result, err = instance1.LocalFindValue(key)
	if result != nil && err == nil {
		count++
	}
	//Within all 22 nodes
	//the number of nodes that store the value should be 20
	if count != 20 {
		t.Error("DoIterativeStore Failed")
	}
}

// EXTRACREDIT
//Check out the Correctness of DoIterativeFindValue
func TestIterativeFindValue(t *testing.T) {
	// tree structure;
	// A->B->tree->tree2
	/*
		                F
			  /
		          C --G
		         /    \
		       /        H
		   A-B -- D
		       \
		          E
	*/

	instance1 := NewKademlia("localhost:7406")
	instance2 := NewKademlia("localhost:7407")
	host2, port2, _ := StringToIpPort("localhost:7407")
	instance1.DoPing(host2, port2)

	//Build the  A->B->Tree structure
	tree_node := make([]*Kademlia, 20)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(7408+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}
	//Build the A->B->Tree->Tree2 structure
	tree_node2 := make([]*Kademlia, 20)
	for j := 20; j < 40; j++ {
		address := "localhost:" + strconv.Itoa(7408+j)
		tree_node2[j-20] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		for i := 0; i < 20; i++ {
			tree_node[i].DoPing(host_number, port_number)
		}
	}

	//Store value into nodes
	value := []byte("Hello world")
	key := NewRandomID()
	contacts, err := instance1.DoIterativeStore(key, value)
	if err != nil || len(contacts) != 20 {
		t.Error("Error doing DoIterativeStore")
	}

	//After Store, check out the correctness of DoIterativeFindValue
	result, err := instance1.DoIterativeFindValue(key)
	if err != nil || result == nil {
		t.Error("Error doing DoIterativeFindValue")
	}

	//Check the correctness of the value we find
	res := string(result[:])
	fmt.Println(res)
	//t.Error("Finish")
}
