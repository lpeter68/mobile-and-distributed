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
	
	go Node.ReceiveMessage("8000")

	rt2 := NewRoutingTable(NewContact(NewKademliaID("1FFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	Node2 := NewKademlia(*rt2, 20, 3 )

	go Node2.ReceiveMessage("8001")

	Node.PingContact(&Node2.routingTable.me)

	time.Sleep(3000 * time.Millisecond)						
	
	contacts := Node.routingTable.FindClosestContacts(NewKademliaID("1FFFFFFF00000000000000000000000000000000"),1)
	contacts2 := Node2.routingTable.FindClosestContacts(NewKademliaID("FFFFFFFF00000000000000000000000000000000"),1)
	
	if !(len(contacts)==1 && contacts[0].ID.String()=="1FFFFFFF00000000000000000000000000000000"){
		t.Error("contact not add after the ping ")		
	}

	if !(len(contacts2)==1 && contacts2[0].ID.String()=="FFFFFFFF00000000000000000000000000000000") {
		t.Error("contact not add after the ping ")		
	}
	fmt.Println("-----PASS")
}

func TestNodeJoin(t *testing.T) {
	fmt.Println("------TestNodeJoin")	
	rt := NewRoutingTable(NewContact(NewKademliaID("1FFFFFFF00000000000000000000000000000000"), "localhost:9000"))
	Node := NewKademlia(*rt, 20, 3 )
	
	go Node.ReceiveMessage("9000")

	rt2 := NewRoutingTable(NewContact(NewKademliaID("2FFFFFFF00000000000000000000000000000000"), "localhost:9001"))
	Node2 := NewKademlia(*rt2, 20, 3 )

	go Node2.ReceiveMessage("9001")
	
	rt3 := NewRoutingTable(NewContact(NewKademliaID("3FFFFFFF00000000000000000000000000000000"), "localhost:9002"))
	Node3 := NewKademlia(*rt3, 20, 3 )

	go Node3.ReceiveMessage("9002")

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
