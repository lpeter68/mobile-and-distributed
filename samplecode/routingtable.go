package main

import (
	"sync"
	//"fmt"
)

const bucketSize = 20

type RoutingTable struct {
	me      Contact
	buckets [IDLength * 8]*bucket
	mutexRT sync.Mutex
}

func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.me = me
	return routingTable
}

func (routingTable *RoutingTable) AddContact(contact Contact, network *Network) {
	//fmt.Println("Add")
	if(contact.ID.String()!=routingTable.me.ID.String()){
		routingTable.mutexRT.Lock()		
		bucketIndex := routingTable.getBucketIndex(contact.ID)
		bucket := routingTable.buckets[bucketIndex]
		bucket.AddContact(contact, routingTable.me, network)
		routingTable.mutexRT.Unlock()		
	}
}

func (routingTable *RoutingTable) RemoveContact(contact Contact) {
	//fmt.Println("Remove")
	if(contact.ID.String()!=routingTable.me.ID.String()){
		routingTable.mutexRT.Lock()		
		bucketIndex := routingTable.getBucketIndex(contact.ID)
		bucket := routingTable.buckets[bucketIndex]
		bucket.RemoveContact(contact, routingTable.me)
		routingTable.mutexRT.Unlock()		
	}
}

func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	routingTable.mutexRT.Lock()
	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	routingTable.mutexRT.Unlock()
	return candidates.GetContacts(count)
}

func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}
