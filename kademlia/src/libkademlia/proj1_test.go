package libkademlia

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	//"fmt"
	//"time"
)

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	instance1 := NewKademlia("localhost:7890")
	instance2 := NewKademlia("localhost:7891")
	host2, port2, _ := StringToIpPort("localhost:7891")
	contact2, err := instance2.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("A node cannot find itself's contact info")
	}
	contact2, err = instance2.FindContact(instance1.NodeID)
	if err == nil {
		t.Error("Instance 2 should not be able to find instance " +
			"1 in its buckets before ping instance 1")
	}
	instance1.DoPing(host2, port2)

	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	contact2, err = instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	wrong_ID := NewRandomID()
	_, err = instance2.FindContact(wrong_ID)
	if err == nil {
		t.Error("Instance 2 should not be able to find a node with the wrong ID")
	}

	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}

	return
}

func TestStore(t *testing.T) {
	// test Dostore() function and LocalFindValue() function
	instance1 := NewKademlia("localhost:7892")
	instance2 := NewKademlia("localhost:7893")
	host2, port2, _ := StringToIpPort("localhost:7893")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	key := NewRandomID()
	value := []byte("Hello World")
	err = instance1.DoStore(contact2, key, value)
	if err != nil {
		t.Error("Can not store this value")
	}
	storedValue, err := instance2.LocalFindValue(key)
	if err != nil {
		t.Error("Stored value not found!")
	}
	if !bytes.Equal(storedValue, value) {
		t.Error("Stored value did not match found value")
	}

	return
}

//Check whether Contact1 and Contact2 are the same Contact
func isSameContact(c1 *Contact, c2 *Contact) bool {
	if c2.NodeID.Equals(c1.NodeID) && c2.Port == c1.Port && c2.Host.Equal(c1.Host) {
		return true
	}
	return false
}

func TestFindNode(t *testing.T) {
	// tree structure;
	// A->B->tree
	/*
	         C
	      /
	  A-B -- D
	      \
	         E
	*/
	instance1 := NewKademlia("localhost:7894")
	instance2 := NewKademlia("localhost:7895")
	host2, port2, _ := StringToIpPort("localhost:7895")
	instance1.DoPing(host2, port2)

	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}

	tree_node := make([]*Kademlia, 10)
	for i := 0; i < 10; i++ {
		address := "localhost:" + strconv.Itoa(7896+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}

	key := NewRandomID()
	contacts, err := instance1.DoFindNode(contact2, key)

	if err != nil {
		t.Error("Error doing FindNode")
	}

	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}

	// TODO: Check that the correct contacts were stored
	//       (and no other contacts)

	// EXTRACREDIT
	//check we found all node correctly when total node smaller than 20
	// check each one in tree_node is in instance1's contact list
	for i := 0; i < 10; i++ {
		if c, err := instance2.FindContact(tree_node[i].SelfContact.NodeID); err != nil && isSameContact(c, &(tree_node[i].SelfContact)) {
			t.Error("Error finding contact ")
		}
	}
	return
}

// EXTRACREDIT
//Check that we found out 20 closest node correctly
//when total node number bigger than 20
func testFindNode2(t *testing.T) {
	instance1 := NewKademlia("localhost:7894")
	instance2 := NewKademlia("localhost:7895")

	host2, port2, _ := StringToIpPort("localhost:7895")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}

	key := NewRandomID()

	minPrefix := instance1.SelfContact.NodeID.Xor(key).PrefixLen()
	// total number of nodes is 21
	tree_node := make([]*Kademlia, 21)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(7896+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
		//get the distance of the farest node
		if prefix := tree_node[i].SelfContact.NodeID.Xor(key).PrefixLen(); prefix < minPrefix {
			minPrefix = prefix
		}
	}

	tree_node[20] = instance1

	contacts, err := instance1.DoFindNode(contact2, key)

	if err != nil {
		t.Error("Error doing FindNode")
	}
	//check we actually find 20 node
	if contacts == nil || len(contacts) != 20 {
		t.Error("Number of Contact is incorrect")
	}

	// check the node in tree_nodes but not in returned contact is the farest one
	for i := 0; i < 21; i++ {
		found := false
		for j := 0; j < 20; j++ {
			if isSameContact(&(tree_node[i].SelfContact), &(contacts[j])) {
				found = true
				break
			}
		}
		if !found {
			//check the node we not contact is the farest one
			if tree_node[i].SelfContact.NodeID.Xor(key).PrefixLen() == minPrefix {
				break
			} else {
				t.Error("Error finding closest nodes")
			}
		}
	}
}

func TestFindValue(t *testing.T) {
	// tree structure;
	// A->B->tree
	/*
	         C
	      /
	  A-B -- D
	      \
	         E
	*/
	instance1 := NewKademlia("localhost:7926")
	instance2 := NewKademlia("localhost:7927")
	host2, port2, _ := StringToIpPort("localhost:7927")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}

	tree_node := make([]*Kademlia, 10)
	for i := 0; i < 10; i++ {
		address := "localhost:" + strconv.Itoa(7928+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}

	key := NewRandomID()
	value := []byte("Hello world")
	err = instance2.DoStore(contact2, key, value)
	if err != nil {
		t.Error("Could not store value")
	}

	// Given the right keyID, it should return the value
	foundValue, contacts, err := instance1.DoFindValue(contact2, key)
	if !bytes.Equal(foundValue, value) {
		t.Error("Stored value did not match found value")
	}

	//Given the wrong keyID, it should return k nodes.
	wrongKey := NewRandomID()
	foundValue, contacts, err = instance1.DoFindValue(contact2, wrongKey)
	if contacts == nil || len(contacts) < 10 {
		t.Error("Searching for a wrong ID did not return contacts")
	}

	return
	// TODO: Check that the correct contacts were stored
	//       (and no other contacts)
	// check the returned contacts are the same as those in tree_node
	for i := 0; i < 10; i++ {
		if c, err := instance2.FindContact(tree_node[i].SelfContact.NodeID); err != nil && isSameContact(c, &(tree_node[i].SelfContact)) {
			t.Error("Error finding contact ")
		}
	}
}
