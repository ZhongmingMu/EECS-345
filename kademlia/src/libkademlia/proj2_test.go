package libkademlia

import (
	//"bytes"
	"fmt"
	//"net"
	"strconv"
	"testing"
	//"container/heap"
	"time"
)


func TestIterativeFindNode3(t *testing.T) {
	// tree structure;
	// A->B->tree（1 child -> 2 child -> 3 child)
	time.Sleep(300 * time.Millisecond)
	instance1 := NewKademlia("localhost:7950")
	instance2 := NewKademlia("localhost:7951")
	host2, port2, _ := StringToIpPort("localhost:7951")
	instance1.DoPing(host2, port2)
	_, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}

	tree_node := make([]*Kademlia,2)
	nodes1 := make([]*Kademlia, 4)
	nodes2 := make([]*Kademlia, 3)
	num := 0

	for i := 0; i < 2; i++ {
		address := "localhost:" + strconv.Itoa(7952+num)
		num++
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)

		for j := 0; j < 2; j++ {
			address_v := "localhost:" + strconv.Itoa(7952+num)

			num++
			nodes1[i * 2 + j] = NewKademlia(address_v)
			host_number_v, port_number_v, _ := StringToIpPort(address_v)
			tree_node[i].DoPing(host_number_v, port_number_v)
			for m := 0; m < 3; m++ {
				address_r := "localhost:" + strconv.Itoa(7952+num)
				num++
				nodes2[m] = NewKademlia(address_r)
				host_number_r, port_number_r, _ := StringToIpPort(address_r)
				nodes1[i * 2 + j].DoPing(host_number_r, port_number_r)
			}
		}
	}

	//key := NewRandomID()
	contacts, err := instance1.DoIterativeFindNode(nodes2[0].NodeID)

	//fmt.Println(key)
	fmt.Println(instance2.SelfContact.NodeID)
	fmt.Println("instance2 up")
	if err != nil {
		t.Error("Error doing FindNode")
	}

	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}
	for _,result := range tree_node {
		fmt.Println(result.NodeID)
	}
	fmt.Println()
	for _,result := range nodes1 {
		fmt.Println(result.NodeID)
	}
	fmt.Println()
	for _,result := range nodes2 {
		fmt.Println(result.NodeID)
	}
	fmt.Println()
	for _,result := range contacts {
		fmt.Println(result.NodeID)
	}
	if len(contacts) != 10 {
		t.Error(len(contacts))
	}
	return
}
