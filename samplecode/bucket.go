package main

import (
	"container/list"
	"fmt"
)

type bucket struct {
	list *list.List
}

func newBucket() *bucket {
	bucket := &bucket{}
	bucket.list = list.New()
	return bucket
}

func (bucket *bucket) AddContact(contact Contact, sourceContact Contact, network *Network) {
	//fmt.Println("AddContact")
	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}
	if element == nil {
		if bucket.list.Len() < bucketSize {
			//fmt.Println("bucket have place")
			bucket.list.PushFront(contact)
		}else{ //Test if the oldest node still alive
			//fmt.Println("bucket full")
			//network.SendPingMessage(sourceContact,bucket.list.Back().Value.(Contact))
				//fmt.Println("oldest node is dead")
				bucket.list.Remove(bucket.list.Back())
				bucket.list.PushFront(contact)
			
		}
	} else {
		//fmt.Println("Allready exist")
		bucket.list.MoveToFront(element)
	}
}


func (bucket *bucket) RemoveContact(contact Contact, sourceContact Contact) {
	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}
	if element != nil {		
		bucket.list.Remove(element)	
		fmt.Println("remove contact")
	}
}


func (bucket *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

func (bucket *bucket) Len() int {
	return bucket.list.Len()
}
