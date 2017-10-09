package main

import (
	"fmt"
	"time"
	"testing"
)

func TestPingNode(t *testing.T) {
	fmt.Println("------TestPingNode")
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))
	Node := NewKademlia(*rt, 20, 3 )

	rt2 := NewRoutingTable(NewContact(NewKademliaID("1FFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	Node2 := NewKademlia(*rt2, 20, 3 )

	time.Sleep(3000 * time.Millisecond)

	Node.PingContact(&Node2.routingTable.me)

	time.Sleep(3000 * time.Millisecond)

	contacts := Node.routingTable.FindClosestContacts(NewKademliaID("1FFFFFFF00000000000000000000000000000000"),1)
	contacts2 := Node2.routingTable.FindClosestContacts(NewKademliaID("FFFFFFFF00000000000000000000000000000000"),1)

	if !(len(contacts)==1 && contacts[0].ID.String()=="1FFFFFFF00000000000000000000000000000000"){
		t.Error("contact not added after the ping ")
	}

	if !(len(contacts2)==1 && contacts2[0].ID.String()=="FFFFFFFF00000000000000000000000000000000") {
		t.Error("contact not added after the ping ")
	}
	fmt.Println("-----PASS")
}

func TestNodeJoin(t *testing.T) {
	fmt.Println("------TestNodeJoin")
	rt := NewRoutingTable(NewContact(NewKademliaID("1FFFFFFF00000000000000000000000000000000"), "localhost:9000"))
	Node := NewKademlia(*rt, 20, 3 )

	rt2 := NewRoutingTable(NewContact(NewKademliaID("2FFFFFFF00000000000000000000000000000000"), "localhost:9001"))
	Node2 := NewKademlia(*rt2, 20, 3 )

	rt3 := NewRoutingTable(NewContact(NewKademliaID("3FFFFFFF00000000000000000000000000000000"), "localhost:9002"))
	Node3 := NewKademlia(*rt3, 20, 3 )

	time.Sleep(3000 * time.Millisecond)

	Node.PingContact(&Node2.routingTable.me)

	time.Sleep(3000 * time.Millisecond)

	contactsBeforejoin := Node2.routingTable.FindClosestContacts(NewKademliaID("3FFFFFFF00000000000000000000000000000000"),1)

	if !(len(contactsBeforejoin)==1 && contactsBeforejoin[0].ID.String()=="1FFFFFFF00000000000000000000000000000000"){
		t.Error("contact not add after the ping ")
	}

	Node3.JoinNetwork(Node.routingTable.me)

	time.Sleep(3000 * time.Millisecond)

	contactsAfterjoin := Node2.routingTable.FindClosestContacts(NewKademliaID("3FFFFFFF00000000000000000000000000000000"),1)

	if !(len(contactsAfterjoin)==1 && contactsAfterjoin[0].ID.String()=="3FFFFFFF00000000000000000000000000000000"){
		t.Error("contact not add after the join ")
	}
	fmt.Println("-----PASS")
}

func testLink(t *testing.T)  {
	/*IDs := []string{
		"FFFFFFFF00000000000000000000000000000000",
		"FFFFFFFF0000000000000000000000FFFFFFFFFF",
		"FFFFFFFF00000000000FFFFFFFFFFFFFFFFFFFFF",
		"FFFFFFFF00000011111111111111111111111111",
		"0000000000000011111111111111111111111111",
		"0000000000000000000000000000000222222222",
	}

	addresses := []string{
		"127.0.0.1:8000",
		"127.0.0.1:8001",
		"127.0.0.1:8002",
		"127.0.0.1:8003",
		"127.0.0.1:8004",
		"127.0.0.1:8005",
	}

	nodes := []string{
		"a1",
		"b1",
		"c1",
		"d1",
		"e1",
		"f1",
	}

	nodes1 := []string{
		"a1",
		"a1",
		"a1",
		"a1",
		"a1",
		"a1",
		"b1",
		"b1",
		"b1",
		"b1",
		"c1",
		"c1",
		"d1",
		"d1",
		"e1"
	}
	nodes2 := []string{
		"b1",
		"c1",
		"d1",
		"e1",
		"f1",
		"c1",
		"d1",
		"e1",
		"f1",
		"d1",
		"e1",
		"f1",
		"e1",
		"f1",
		"f1",
	}

	mapKademlia := make(map[string]*Kademlia)

	for i := 0; i < len(IDs); i++ {
		rt := NewRoutingTable(NewContact(NewKademliaID(IDs[i]), addresses[i]))
		mapKademlia[nodes[i]] = NewKademlia(*rt, 20, 3)
	}

	for k := 0; k < len(IDs); k++ {
		for j := 0; j < len(nodes1); i++ {
			mapKademlia[nodes1[j]].PingContact(&mapKademlia[nodes2[j]].routingTable.me)
			mapKademlia[nodes2[j]].PingContact(&mapKademlia[nodes1[j]].routingTable.me)
			contacts := rt.FindClosestContacts(IDs[k], 20)

		}
	}*/
	fmt.Println("------TestLinkNode")
	mapKademlia := make(map[string]*Kademlia)
	found := false
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "127.0.0.1:8000"))
	kademlia1 := NewKademlia(*rt, 20, 3)
	mapKademlia["a1"] = kademlia1
	rt2 := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF0000000000000000000000FFFFFFFFFF"), "127.0.0.1:8001"))
	kademlia2 := NewKademlia(*rt2, 20, 3)
	mapKademlia["b1"] = kademlia2


	mapKademlia["a1"].PingContact(&mapKademlia["b1"].routingTable.me)
	mapKademlia["b1"].PingContact(&mapKademlia["a1"].routingTable.me)

	contacts1 := rt.FindClosestContacts(NewKademliaID("FFFFFFFF0000000000000000000000FFFFFFFFFF"), 20)
	contacts2 := rt2.FindClosestContacts(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), 20)
	for i := 0; i < len(contacts1); i++ {
		if (contacts1[i].ID.String()=="FFFFFFFF0000000000000000000000FFFFFFFFFF" )
			found = true
	}

	for i := 0; i < len(contacts2); i++ {
		if (contacts2[i].ID.String()=="FFFFFFFF00000000000000000000000000000000" && found==true )
			fmt.Println("-----PASS")
	}

}

func storeFIleTest(t *testing.T)  {
	fmt.Println("------TestLinkNode and LookUpData")
	mapKademlia := make(map[string]*Kademlia)
	rt := NewRoutingTable(NewContact(NewHashKademliaId("test.txt"), "127.0.0.1:8000"))
	kademlia1 := NewKademlia(*rt, 20, 3)
	mapKademlia["a1"] = kademlia1
	fileName := "test.txt"
	file, err := os.Open(fileName)
	CheckError(err)
	stat,_ := file.Stat()
	size := stat.Size()
	var fileBuffer []byte = make([]byte, size)
	_, err := file.Read(fileBuffer)
	if err != io.EOF {
			mapKademlia["a1"].Store(&File{fileName, fileBuffer})
			/*It would be great to add a var in lookupData to know if it's found or not */
			mapKademlia["a1"].LookupData("test.txt")
	}
}
