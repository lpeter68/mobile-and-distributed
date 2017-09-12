package main

import "fmt"
import "bufio"
//import "os"
import "encoding/json"
import "net"


type Kademlia struct {
	routingTable RoutingTable
	k int
	alpha int
	network Network
	data map[string]File //map kademlia id to data
}

func NewKademlia(rt RoutingTable, k int, alpha int) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.routingTable = rt
	kademlia.k = k
	kademlia.alpha = alpha
	kademlia.network = Network{}
	kademlia.data = make(map[string]File) 	
	return kademlia
}

func (kademlia *Kademlia) PingContact(target *Contact) bool{
	result := SendPingMessage(kademlia.routingTable.me,*target)
	if(result){
		kademlia.routingTable.AddContact(*target)
	}
	return result
}

func (kademlia *Kademlia) LookupContact(target *Contact) []Contact{ //TODO change param to kademliaId type
	var closestContact map[string]*Contact = make(map[string]*Contact) //map kademlia id to contact
	var alreadyLookup map[string]bool = make(map[string]bool) //true if allready lookup
	var nextLookup []Contact = kademlia.routingTable.FindClosestContacts(target.ID,kademlia.alpha)
	addToMap(&closestContact,&alreadyLookup,nextLookup,target)
	/*for i := range nextLookup {
		fmt.Println("Lookup 0: " +string(i) +" : "+ nextLookup[i].ID.String())	
	}*/
	for i := 0; i <= kademlia.k; i++ {
		for j := 0; j<len(nextLookup); i, j = i+1, j+1 {
			closestContactFind := SendFindContactMessage(kademlia.routingTable.me, nextLookup[j],*target)
			alreadyLookup[nextLookup[j].ID.String()]=true		
			addToMap(&closestContact,&alreadyLookup,closestContactFind,target)
		}
		nextLookup = kademlia.findNextLookup(&closestContact,&alreadyLookup,target, false)
		/*fmt.Println("Lookup n°" +string(i) +" : ")			
		for i := range nextLookup {
			fmt.Println("ID n°"+string(i)+" : "+ nextLookup[i].ID.String())	
		}
		fmt.Println("	Map n°" +string(i) +" : ")			
		for i := range closestContact {
			if(alreadyLookup[i]){
				fmt.Print(i + " state : true ")
			}else{
				fmt.Print(i + " state : false ")				
			}
			if(closestContact[i].distance!=nil){
				fmt.Println(closestContact[i].distance.String())	
			}
		}*/
	}
	endLookup := kademlia.findNextLookup(&closestContact,&alreadyLookup,target, true)
	fmt.Println(" k closest contact find : ")			
	for i := range endLookup {
		fmt.Println("ID n°"+string(i)+" : "+ endLookup[i].ID.String())	
	}
	return endLookup
}

func (kademlia *Kademlia) findNextLookup(mpContact *map[string]*Contact, mpBool *map[string]bool , target *Contact, finalLookup bool) []Contact {
	var size int
	if(finalLookup){
		size=kademlia.k
	}else{
		size=kademlia.alpha	
	}
	mContact := *mpContact
	mBool := *mpBool	
	result := make([]Contact, size)
	nextEmptyIndex := 0;
	for i := range mContact {
		mContact[i].CalcDistance(target.ID) 
		if(!mBool[i] || finalLookup){
			if(nextEmptyIndex<size){
				result[nextEmptyIndex]=*mContact[i]
				nextEmptyIndex++
			}else{
				var indexMax int = 0
				var contactMax Contact = result[0]
				for j := 1; j<len(result); j++ {
					if(contactMax.Less(&result[j])){
						contactMax = result[j]
						indexMax = j
					}
				}
				if(mContact[i].Less(&contactMax)){
					result[indexMax] = *mContact[i]
				}
			}
		}
	}
    return result
}

func addToMap(mpContact *map[string]*Contact, mpBool *map[string]bool, contacts []Contact, target *Contact) {
	mContact := *mpContact
	mBool := *mpBool
	for i := range contacts {
		contacts[i].CalcDistance(target.ID)
		_, exist := mContact[contacts[i].ID.String()]
		if(!exist){
			//fmt.Println("Add : "+ contacts[i].ID.String())			
			mContact[contacts[i].ID.String()] = &contacts[i]
			mBool[contacts[i].ID.String()] = false			
		}
	}
}

func (kademlia *Kademlia) LookupData(title string) []byte{
	target := NewContact(NewHashKademliaId(title),"")
	var closestContact map[string]*Contact = make(map[string]*Contact) //map kademlia id to contact
	var alreadyLookup map[string]bool = make(map[string]bool) //true if allready lookup
	var nextLookup []Contact = kademlia.routingTable.FindClosestContacts(target.ID,kademlia.alpha)
	addToMap(&closestContact,&alreadyLookup,nextLookup,&target)

	for i := 0; i <= kademlia.k; i++ {
		for j := 0; j<len(nextLookup); i, j = i+1, j+1 {
			data := SendFindDataMessage(kademlia.routingTable.me, nextLookup[j],title)
			if(data!=nil){
				fmt.Println("Data found on node : "+nextLookup[j].ID.String())							
				return data
			}
			closestContactFind := SendFindContactMessage(kademlia.routingTable.me, nextLookup[j],target)
			alreadyLookup[nextLookup[j].ID.String()]=true		
			addToMap(&closestContact,&alreadyLookup,closestContactFind,&target)
		}
		nextLookup = kademlia.findNextLookup(&closestContact,&alreadyLookup,&target, false)
	}
	fmt.Println("Data not found ")			
	return nil
}

func (kademlia *Kademlia) Store(file File) {
	fileContact := NewContact(NewHashKademliaId(file.Title),"")
	closestNodes := kademlia.LookupContact(&fileContact)
	for i := range closestNodes {
		SendStoreMessage(kademlia.routingTable.me,closestNodes[i],file)	
	}
}

func (kademlia *Kademlia) AddToNetwork(adress string) {
	closestNodes := SendAddNodeMessage(kademlia.routingTable.me,adress)
	for i := range closestNodes {
		kademlia.PingContact(&closestNodes[i])	
	}
}

func (kademlia *Kademlia) ReceiveMessage(port string) {
	//fmt.Println("Launching server...")
	
	  // listen on all interfaces
	  ln, _ := net.Listen("tcp", ":"+port)
	 for{
	  	// accept connection on port
	  	conn, _ := ln.Accept()
	
		// will listen for message to process ending in newline (\n)
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// output message received
		//fmt.Print("Message Received:", string(message))
		var messageDecoded Message

		json.Unmarshal([]byte(message),&messageDecoded)
		//fmt.Println("Message type Received:", messageDecoded.MessageType)	
		var responseMessage Message 		
		switch(messageDecoded.MessageType){
			case PING :
				//fmt.Println("Message Ping Received:", string(messageDecoded.Content[0]))
				go kademlia.routingTable.AddContact(messageDecoded.Source)
				responseMessage = Message{kademlia.routingTable.me, RESPONSE , ""}
			break

			case FINDCONTACT :
				//fmt.Println("Message findContact Received:", string(messageDecoded.Content[0]))	
				closestContact := kademlia.routingTable.FindClosestContacts(NewKademliaID(messageDecoded.Content),kademlia.alpha)
				JSONClosestContact, _ := json.Marshal(closestContact)
				responseMessage = Message{kademlia.routingTable.me, RESPONSE , string(JSONClosestContact)}
			break

			case FINDDATA :
				//fmt.Println("Message findData Received:", string(messageDecoded.Content[0]))
				_, exist := kademlia.data[messageDecoded.Content]
				if(exist){
					JSONData, _ := json.Marshal(kademlia.data[messageDecoded.Content])
					responseMessage = Message{kademlia.routingTable.me, RESPONSE , string(JSONData)}
				}else{
					responseMessage = Message{kademlia.routingTable.me, RESPONSE , ""}				
				}
			break

			case STORE :
				//fmt.Println("Message store Received:", string(messageDecoded.Content[0]))
				var dataDecoded File			
				json.Unmarshal([]byte(messageDecoded.Content),&dataDecoded)
				kademlia.data[NewHashKademliaId(dataDecoded.Title).String()]=dataDecoded
				responseMessage = Message{kademlia.routingTable.me, RESPONSE , ""}
			break

			case ADDNODE :
				//fmt.Println("Message addNode Received:", string(messageDecoded.Content[0]))
				closestContact := kademlia.LookupContact(&messageDecoded.Source)
				JSONClosestContact, _ := json.Marshal(closestContact)
				responseMessage = Message{kademlia.routingTable.me, RESPONSE , string(JSONClosestContact)}
			break
									
			default :
				fmt.Println("Unexpected Message Received:", string(message))	
			break				
		}

		JSONResponseMessage, _ := json.Marshal(responseMessage)

		// sample process for string received
		//var a []byte = []byte("Response \n");

		//fmt.Print("message to byte", string(JSONResponseMessage))		
		conn.Write([]byte(string(JSONResponseMessage) +"\n"))
		//ln.Close();
	 }
}

