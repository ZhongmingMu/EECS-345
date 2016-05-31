package libkademlia
//
// import (
// 	//"bytes"
// 	//"fmt"
// 	//"net"
// 	"strconv"
// 	"testing"
// 	//"container/heap"
// //	"time"
// )
//
// // EXTRACREDIT
// // Check out the correctness of DoIterativeFindNode function
// func TestIterativeFindNode1(t *testing.T) {
// 	// tree structure;
// 	// A<--->B
// tree_node := make([]*Kademlia, 30)
// address := make([]string, 30)
// for i := 0; i < 30; i++ {
// 	address[i] = "localhost:" + strconv.Itoa(7000+i)
// 	tree_node[i] = NewKademlia(address[i])
// }
//
// //30 nodes ping each other
// for i := 0; i < 30 ; i++ {
// 	for j := 0; j < 30; j++ {
// 		host_number, port_number, _ := StringToIpPort(address[j])
// 		tree_node[i].DoPing(host_number, port_number)
// 	}
// }
//
// //find node[19], start from node 0
// contacts, _ := tree_node[0].DoIterativeFindNode(tree_node[19].SelfContact.NodeID)
// count := 0
// //check the result
// for i := 0; i < len(contacts); i++ {
// 	if(contacts[i].NodeID.Equals(tree_node[19].SelfContact.NodeID)) {
// 		count ++
// 	}
// 	//fmt.Print(contacts[i].NodeID)
// }
// if(count != 1) {
// 	t.Error("the result is not true")
// }
// return
// }
//
// // EXTRACREDIT
// // Check out the correctness of DoIterativeFindNode function
// func TestIterativeFindNode2(t *testing.T) {
// 	// tree structure;
// 	// A<--->B
// 	total_num := 200											//total num mush >= 20
// 	tree_node := make([]*Kademlia, total_num)
// 	instance1 := NewKademlia("localhost:7399")										//starting node
// 	host1, port1, _ := StringToIpPort("localhost:7399")
//
// 	findId := total_num - 20
// 	//initialize the nodes
// 	for i := 0; i < total_num; i++ {
// 		address := "localhost:" + strconv.Itoa(7400+i)
// 		tree_node[i] = NewKademlia(address)
// 		tree_node[i].DoPing(host1, port1)														//every node ping instance1
// 	}
// 	target := tree_node[findId].SelfContact.NodeID								//id to be found
//
// 	//node less that findId ping the findId, not all ping the findId node
// 	for i := 0; i < findId; i++ {
// 			tree_node[findId].DoPing(tree_node[i].SelfContact.Host, tree_node[i].SelfContact.Port)
// 	}
//
// 	result, err := instance1.DoIterativeFindNode(target)						//start from instance1,find target
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
//
// 	if result == nil || len(result) == 0 {
// 		t.Error("No contacts were found")
// 	}
//
// 	//check the result
// 	count := 0
// 	for _, value := range result {
// 		if value.NodeID.Equals(target) {
// 			count++
// 		}
// 	}
// 	//if only Contains one same node, test pass
// 	if count != 1 {
// 		t.Error(count)
// 	}
// 	return
// }
// // EXTRACREDIT
// // Check out the correctness of DoIterativeFindNode function
// func TestIterativeFindNode3(t *testing.T) {
// 	// tree structure;
// 	//  A--> B --> 5 nodes --> 5 nodes --> 5nodes
// 	/*
// 		                F
// 			             /
// 		          C --G    J
// 		         /    \  /
// 		       /        H --I
// 		   A-B -- D      \
// 		       \          K
// 		        E
// */
// 	instance1 := NewKademlia("localhost:4480")
// 	instance2 := NewKademlia("localhost:4481")
// 	host2, port2, _ := StringToIpPort("localhost:4481")
// 	instance1.DoPing(host2, port2)
// 	_, err := instance1.FindContact(instance2.NodeID)
// 	if err != nil {
// 		t.Error("Instance 2's contact not found in Instance 1's contact list")
// 		return
// 	}
//   //Build the Tree Structure
// 	tree_node := make([]*Kademlia,5)
// 	count := 0
// 	for i := 0; i < 5; i++ {
// 		address := "localhost:" + strconv.Itoa(4482+count)
// 		count++
// 		tree_node[i] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		instance2.DoPing(host_number, port_number)
//
// 		for j := 0; j < 5; j++ {
// 			address_v := "localhost:" + strconv.Itoa(4482+count)
//
// 			count++
// 			instance_temp := NewKademlia(address_v)
// 			host_number_v, port_number_v, _ := StringToIpPort(address_v)
// 			tree_node[i].DoPing(host_number_v, port_number_v)
// 			for m := 0; m < 5; m++ {
// 				address_r := "localhost:" + strconv.Itoa(4482+count)
// 				count++
// 				 NewKademlia(address_r)
// 				host_number_r, port_number_r, _ := StringToIpPort(address_r)
// 				instance_temp.DoPing(host_number_r, port_number_r)
// 			}
// 		}
//
// 	}
// 	//Implement DoIterativeFindNode
// 	key := NewRandomID()
// 	contacts, err := instance1.DoIterativeFindNode(key)
// 	//check the result of DoIterativeFindNode
// 	if err != nil || len(contacts) != 20{
// 		t.Error("Error doing DoIterativeFindNode")
// 	}
// 	return
// }
// // EXTRACREDIT
// //Check out the Correctness of DoIterativeStore
// func TestIterativeStore(t *testing.T) {
// 	// tree structure;
// 	// A->B->tree->tree2
// 	/*
// 	          C
// 	      /
// 	   A-B -- D
// 	       \
// 	          E
// */
// 	instance1 := NewKademlia("localhost:6506")
// 	instance2 := NewKademlia("localhost:6507")
// 	host2, port2, _ := StringToIpPort("localhost:6507")
// 	instance1.DoPing(host2, port2)
//
// 	//Build the  A->B->Tree structure
// 	tree_node := make([]*Kademlia, 30)
// 	for i := 0; i < 30; i++ {
// 		address := "localhost:" + strconv.Itoa(6508+i)
// 		tree_node[i] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		instance2.DoPing(host_number, port_number)
// 	}
// 	//implement DoIterativeStore, and get the the result
// 	value := []byte("Hello world")
// 	key := NewRandomID()
// 	contacts, err := instance1.DoIterativeStore(key, value)
// 	//the number of contacts store the value should be 20
// 	if err != nil || len(contacts) != 20 {
// 		t.Error(len(contacts))
// 	}
// 	//Check all the 32 nodes,
// 	//find out the number of nodes that contains the value
// 	count := 0
// 	// check tree_nodes[0~19]
// 	for i := 0; i < 30; i++ {
// 		result, err := tree_node[i].LocalFindValue(key)
// 		if string(result) == string(value) && err == nil {
// 			count++
// 		}
// 	}
// 	//check instance2
// 	result, err := instance2.LocalFindValue(key)
// 	if string(result) == string(value) && err == nil {
// 		count++
// 	}
// 	//check instance1
// 	result, err = instance1.LocalFindValue(key)
// 	if string(result) == string(value) && err == nil {
// 		count++
// 	}
// 	//Within all 32 nodes
// 	//the number of nodes that store the value should be 20
// 	if count != 20 {
// 		t.Error("DoIterativeStore Failed")
// 	}
// }
//
// // EXTRACREDIT
// //Check out the Correctness of DoIterativeFindValue
// //when value stored by DoIterativeStore
// func TestIterativeFindValue1(t *testing.T) {
// 	// tree structure;
// 	// A->B->tree->tree2
// 	/*
// 		                F
// 			  /
// 		          C --G
// 		         /    \
// 		       /        H
// 		   A-B -- D
// 		       \
// 		          E
// */
// 	instance1 := NewKademlia("localhost:5406")
// 	instance2 := NewKademlia("localhost:5407")
// 	host2, port2, _ := StringToIpPort("localhost:5407")
// 	instance1.DoPing(host2, port2)
//
// 	//Build the  A->B->Tree structure
// 	tree_node := make([]*Kademlia, 30)
// 	for i := 0; i < 30; i++ {
// 		address := "localhost:" + strconv.Itoa(5408+i)
// 		tree_node[i] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		instance2.DoPing(host_number, port_number)
// 	}
// 	//Build the A->B->Tree->Tree2 structure
// 	tree_node2 := make([]*Kademlia, 30)
// 	for j := 30; j < 60; j++ {
// 		address := "localhost:" + strconv.Itoa(5408+j)
// 		tree_node2[j-30] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		for i := 0; i < 30; i++ {
// 			tree_node[i].DoPing(host_number, port_number)
// 		}
// 	}
//
// 	//Store value into nodes by DoIterativeStore
// 	value := []byte("Hello world")
// 	key := NewRandomID()
// 	contacts, err := instance1.DoIterativeStore(key, value)
// 	if err != nil || len(contacts) != 20 {
// 		t.Error("Error doing DoIterativeStore")
// 	}
//
// 	//After Store, check out the correctness of DoIterativeFindValue
// 	result, err := instance1.DoIterativeFindValue(key)
// 	//Check the correctness of the value we find
// 	if err != nil || string(result) != string(value) {
// 		t.Error("Error doing DoIterativeFindValue")
// 	}
//
// 	// res := string(result[:])
// 	// fmt.Println(res)
// 	//t.Error("Finish")
// }
//
// // EXTRACREDIT
// //check the correctness of DoIterativeFindNode
// //When value only store in one node
// func TestIterativeFindValue2(t *testing.T) {
// 	// tree structure;
// 	// A->B->tree->tree2
// 	/*
// 	             F
// 	            /
// 	          C —G
// 	         /  \
// 	       /     H
// 	   A-B — D
// 	       \
// 	          E
// */
// 	instance1 := NewKademlia("localhost:5606")
// 	instance2 := NewKademlia("localhost:5607")
// 	host2, port2, _ := StringToIpPort("localhost:5607")
// 	instance1.DoPing(host2, port2)
//
// 	//Build the  A->B->Tree structure
// 	tree_node := make([]*Kademlia, 20)
// 	for i := 0; i < 20; i++ {
// 		address := "localhost:" + strconv.Itoa(5608+i)
// 		tree_node[i] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		instance2.DoPing(host_number, port_number)
// 	}
// 	//Build the A->B->Tree->Tree2 structure
// 	tree_node2 := make([]*Kademlia, 20)
// 	for j := 20; j < 40; j++ {
// 		address := "localhost:" + strconv.Itoa(5608+j)
// 		tree_node2[j-20] = NewKademlia(address)
// 		host_number, port_number, _ := StringToIpPort(address)
// 		for i := 0; i < 20; i++ {
// 			tree_node[i].DoPing(host_number, port_number)
// 		}
// 	}
// 	//Store value in one node of tree2
// 	tmp_contact, err := tree_node[3].FindContact(tree_node2[12].NodeID)
// 	if err != nil {
// 		t.Error("Can't find Contact")
// 	}
// 	value := []byte("Hello world")
// 	err = tree_node[3].DoStore(tmp_contact, tmp_contact.NodeID, value)
// 	if err != nil {
// 		t.Error("Store value failed")
// 	}
//
// 	//After Store, check out the correctness of DoIterativeFindValue
// 	//by using A to find a value in tree2
// 	result, err := instance1.DoIterativeFindValue(tmp_contact.NodeID)
// 	//check out the correctness the value we find
// 	if err != nil || string(result) != string(value)  {
// 		t.Error("Error doing DoIterativeFindValue")
// 	}
//
// 	//t.Error("Finish")
// }
