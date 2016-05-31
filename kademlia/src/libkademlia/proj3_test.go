// Ounan Ma, omg049
// Chong Yan, cyu422
// Wenjie Zhang, wzm416

package libkademlia

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	// "time"
	"fmt"
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

func TestVanish_Unvanish_1(t *testing.T) {
	instance1 := NewKademlia("localhost:8080")
	tree_node_trie1 := make([]*Kademlia, 40)
	for i := 0; i < 40; i++ {
		address := "localhost:" + strconv.Itoa(7081+i)
		tree_node_trie1[i] = NewKademlia(address)
		host1, port1, _ := StringToIpPort(address)
		instance1.DoPing(host1, port1)
	}
	SearchKey := instance1.SelfContact.NodeID
	VdoID := NewRandomID()
	instance1.Vanish(VdoID, []byte("AAAAAA"), 40, 20, 10000000000)
	data := tree_node_trie1[5].Unvanish(SearchKey, VdoID)
	//	fmt.Println(string(ciphertext) + "is result")
	if !bytes.Equal(data, []byte("AAAAAA")) {
		t.Error("Unvanish error")
	}

	return
}

func TestVanish_Unvanish_2(t *testing.T) {
	instance1 := NewKademlia("localhost:5080")
	tree_node_trie1 := make([]*Kademlia, 30)
	tree_node_trie2 := make([]*Kademlia, 20)
	for i := 0; i < 30; i++ {
		address := "localhost:" + strconv.Itoa(3581+i)
		tree_node_trie1[i] = NewKademlia(address)
		host1, port1, _ := StringToIpPort(address)
		instance1.DoPing(host1, port1)
	}
	for j := 0; j < 20; j++ {
		address := "localhost:" + strconv.Itoa(6081+j)
		tree_node_trie2[j] = NewKademlia(address)
		host2, port2, _ := StringToIpPort(address)
		tree_node_trie1[5].DoPing(host2, port2)
	}
	SearchKey := instance1.SelfContact.NodeID
	VdoID := NewRandomID()
	//fmt.Println("test2 ciphertext")
	instance1.Vanish(VdoID, []byte("AAAAAA"), 30, 25, 100000000000)
	fmt.Println("test2 ciphertext")
	ciphertext := tree_node_trie2[3].Unvanish(SearchKey, VdoID)
	fmt.Println("test2 original")
	fmt.Println([]byte("AAAAAA"))
	//	fmt.Println(string(ciphertext) + "is result")
	if !bytes.Equal(ciphertext, []byte("AAAAAA")) {
		t.Error("test2 Unvanish error")
	}
	//t.Error("Finish")
	return
}

func TestVanish_Unvanish_local(t *testing.T) {
	fmt.Println("starting")
	instance1 := NewKademlia("localhost:3080")
	tree_node_trie1 := make([]*Kademlia, 20)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(4581+i)
		tree_node_trie1[i] = NewKademlia(address)
		host1, port1, _ := StringToIpPort(address)
		instance1.DoPing(host1, port1)
	}
	fmt.Println("finished creating")
	SearchKey := instance1.SelfContact.NodeID
	VdoID := NewRandomID()
	instance1.Vanish(VdoID, []byte("AAAAAA"), 20, 20, 3000000000000)
	data := instance1.Unvanish(SearchKey, VdoID)
	//fmt.Println(string(ciphertext) + "is result")
	if !bytes.Equal(data, []byte("AAAAAA")) {
		t.Error("Unvanish error")
	}
	return
}
