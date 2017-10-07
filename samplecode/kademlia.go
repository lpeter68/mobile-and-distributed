package main

import "sync"
import "fmt"

//import "bufio"
import "time"
import "encoding/json"
import "net"
import "bufio"
import 	"strconv"
import 	"strings"

type Kademlia struct {
	mutexLookup    sync.Mutex
	mutexSem       sync.Mutex
	mutexData      sync.Mutex
	mutexFileSend  sync.Mutex	
	semaphore      int
	nbResponse     int
	routingTable   RoutingTable
	timeoutMessages time.Duration
	timeoutFiles time.Duration
	k              int
	alpha          int
	network        Network
	data           map[string]File     //map kademlia id to data
	alreadyLookup  map[string]int      // use for lookup
	closestContact map[string]*Contact // use for lookup
	dataFound      File
	filesend map[string]*File	
}

func NewKademlia(rt RoutingTable, k int, alpha int) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.routingTable = rt
	kademlia.k = k
	kademlia.alpha = alpha
	kademlia.network = Network{}
	kademlia.data = make(map[string]File)
	kademlia.filesend = make(map[string]*File)	
	kademlia.timeoutMessages = 2*time.Second
	kademlia.timeoutFiles = 1*time.Minute
	go kademlia.ReceiveMessageUdp(strings.Split(rt.me.Address, ":")[1])
	go kademlia.ReceiveMessageTcp(strings.Split(rt.me.Address, ":")[1])		
	go kademlia.Keepdata()	
	return kademlia
}

func (kademlia *Kademlia) PingContact(target *Contact) {
	kademlia.network.SendPingMessage(kademlia.routingTable.me, *target)
}

func (kademlia *Kademlia) LookupContact(target *Contact) []Contact { //TODO change param to kademliaId type
	kademlia.mutexLookup.Lock()
	//fmt.Println("Start lookup contact")
	kademlia.closestContact = make(map[string]*Contact) //map kademlia id to contact
	kademlia.alreadyLookup = make(map[string]int)       //true if already lookup
	kademlia.semaphore = 0
	kademlia.nbResponse = 0
	//fmt.Println("closest contact to find")
	var nextLookup []Contact = kademlia.routingTable.FindClosestContacts(target.ID, kademlia.alpha)
	//fmt.Println("closest contact found")
	kademlia.addToMap(&kademlia.closestContact, &kademlia.alreadyLookup, nextLookup, target.ID)
	//fmt.Println("end of init")
	for kademlia.nbResponse < kademlia.k {
		//fmt.Println("loop 1")
		if len(nextLookup) == 0 {
			//fmt.Println("break")
			break
		}
		for j := 0; j < len(nextLookup); j = j + 1 {
			//fmt.Println("loop 2")
			kademlia.network.SendFindContactMessage(kademlia.routingTable.me, nextLookup[j], *target)
			kademlia.mutexSem.Lock()
			kademlia.alreadyLookup[nextLookup[j].ID.String()] = 1
			kademlia.mutexSem.Unlock()
		}
		k := 0
		for kademlia.semaphore == 0 && k < 10 {
			k++
			time.Sleep(300 * time.Millisecond)
			//fmt.Println("wait semaphore")
		}
		kademlia.semaphore--
		nextLookup = kademlia.findNextLookup(&kademlia.closestContact, &kademlia.alreadyLookup, target, false)
		/*kademlia.mutexSem.Lock()
		fmt.Println("goroutine unlock")
		fmt.Println(len(kademlia.closestContact))
		/*fmt.Println("Lookup")
		for i := range nextLookup{
			fmt.Println(nextLookup[i])
		}
		fmt.Println("map")
		for i := range kademlia.closestContact{
			fmt.Println(kademlia.closestContact[i])
			fmt.Println(kademlia.alreadyLookup[i])
		}
		kademlia.mutexSem.Unlock()*/
	}
	endLookup := kademlia.findNextLookup(&kademlia.closestContact, &kademlia.alreadyLookup, target, true)
	/*for i := range endLookup{
		fmt.Println(endLookup[i])
	}*/
	//fmt.Println("End of lookup")
	kademlia.mutexLookup.Unlock()	
	return endLookup
}

func (kademlia *Kademlia) findNextLookup(mpContact *map[string]*Contact, mpBool *map[string]int, target *Contact, finalLookup bool) []Contact {
	kademlia.mutexSem.Lock()
	var size int
	if finalLookup {
		size = kademlia.k
	} else {
		size = kademlia.alpha
	}
	mContact := *mpContact
	mBool := *mpBool
	result := make([]Contact, 0, size)
	nextEmptyIndex := 0
	for i := range mContact {
		mContact[i].CalcDistance(target.ID)
		if mBool[i] == 0 || finalLookup {
			if nextEmptyIndex < size {
				result = append(result, *mContact[i])
				nextEmptyIndex++
			} else {
				var indexMax int = 0
				var contactMax Contact = result[0]
				for j := 1; j < len(result); j++ {
					if contactMax.Less(&result[j]) {
						contactMax = result[j]
						indexMax = j
					}
				}
				if mContact[i].Less(&contactMax) {
					result[indexMax] = *mContact[i]
				}
			}
		}
	}
	kademlia.mutexSem.Unlock()
	return result
}

func (kademlia *Kademlia) addToMap(mpContact *map[string]*Contact, mpBool *map[string]int, contacts []Contact, target *KademliaID) {
	kademlia.mutexSem.Lock()
	mContact := *mpContact
	mBool := *mpBool
	for i := range contacts {
		contacts[i].CalcDistance(target)
		_, exist := mContact[contacts[i].ID.String()]
		if !exist && contacts[i].ID.String() != kademlia.routingTable.me.ID.String() {
			//fmt.Println("Add : "+ contacts[i].ID.String())
			mContact[contacts[i].ID.String()] = &contacts[i]
			mBool[contacts[i].ID.String()] = 0
		} else if contacts[i].ID.String() == kademlia.routingTable.me.ID.String() {
			mContact[contacts[i].ID.String()] = &contacts[i]
			mBool[contacts[i].ID.String()] = 2
		}
	}
	kademlia.mutexSem.Unlock()
}

func (kademlia *Kademlia) LookupData(title string) []byte {
	kademlia.mutexLookup.Lock()						
	//fmt.Println("Begin of LookupData")
	target := NewContact(NewHashKademliaId(title), "")
	kademlia.dataFound = File{"", nil,false,true,time.Now()}
	kademlia.closestContact = make(map[string]*Contact) //map kademlia id to contact
	kademlia.alreadyLookup = make(map[string]int)       //true if allready lookup
	kademlia.semaphore = 0
	kademlia.nbResponse = 0
	//fmt.Println("closest contact to find")
	var nextLookup []Contact = kademlia.routingTable.FindClosestContacts(target.ID, kademlia.alpha)
	//fmt.Println("closest contact found")
	kademlia.addToMap(&kademlia.closestContact, &kademlia.alreadyLookup, nextLookup, target.ID)
	//fmt.Println("end of init")
	for kademlia.nbResponse < kademlia.k {
		//fmt.Println("loop 1")
		if len(nextLookup) == 0 {
			//fmt.Println("break")
			break
		}
		for j := 0; j < len(nextLookup); j = j + 1 {
			//fmt.Println("loop 2")
			kademlia.network.SendFindDataMessage(kademlia.routingTable.me, nextLookup[j], title)
			kademlia.mutexSem.Lock()
			kademlia.alreadyLookup[nextLookup[j].ID.String()] = 1
			kademlia.mutexSem.Unlock()
		}
		for kademlia.semaphore == 0 && kademlia.dataFound.Title == "" {
			time.Sleep(300 * time.Millisecond)
			//fmt.Println("wait semaphore")
		}
		if(kademlia.dataFound.Title != ""){
			kademlia.mutexLookup.Unlock()					
			return kademlia.dataFound.Data
		}
		kademlia.semaphore--
		nextLookup = kademlia.findNextLookup(&kademlia.closestContact, &kademlia.alreadyLookup, &target, false)
		/*kademlia.mutexSem.Lock()
		//fmt.Println("goroutine unlock")
		//fmt.Println(len(kademlia.closestContact))
		/*fmt.Println("Lookup")
		for i := range nextLookup{
			fmt.Println(nextLookup[i])
		}
		fmt.Println("map")
		for i := range kademlia.closestContact{
			fmt.Println(kademlia.closestContact[i])
			fmt.Println(kademlia.alreadyLookup[i])
		}
		kademlia.mutexSem.Unlock()*/
	}
	//endLookup := kademlia.findNextLookup(&kademlia.closestContact,&kademlia.alreadyLookup,&target, true)
	fmt.Println("data not found")
	kademlia.mutexLookup.Unlock()		
	return kademlia.dataFound.Data
}

func (kademlia *Kademlia) Store(file *File) {
	kademlia.mutexFileSend.Lock()
	_, exist := kademlia.filesend[file.Title]
	kademlia.mutexFileSend.Unlock()
	if exist{
		fmt.Println("Impossible to store file already exist")
	}else{
		kademlia.mutexFileSend.Lock()
		kademlia.filesend[file.Title]=file
		kademlia.mutexFileSend.Unlock()
		fileContact := NewContact(NewHashKademliaId(file.Title),"")
		
		for file.on{
			closestNodes := kademlia.LookupContact(&fileContact)
			kademlia.mutexFileSend.Lock()
			kademlia.filesend[file.Title].LastStoreMessage=time.Now()
			kademlia.mutexFileSend.Unlock()
			for i := range closestNodes {
				kademlia.network.SendStoreMessage(kademlia.routingTable.me,closestNodes[i],file)
			}
			time.Sleep((3*kademlia.timeoutFiles)/4)
		}
	}
}

func (kademlia *Kademlia) JoinNetwork(contactOnNetwork Contact) {
	//fmt.Println("Start add to network")
	//kademlia.PingContact(&contactOnNetwork)
	kademlia.routingTable.AddContact(contactOnNetwork, &kademlia.network)
	//fmt.Println("Ping envoyé")
	result := kademlia.LookupContact(&kademlia.routingTable.me)
	for i := range result {
		kademlia.PingContact(&result[i])
	}
}

func (kademlia *Kademlia) Keepdata(){
	for{
		kademlia.mutexData.Lock()		
		for title := range kademlia.data{
			if !kademlia.data[title].on{
				delete(kademlia.data,title)
			}else{
				now:=time.Now()
				if now.Sub(kademlia.data[title].LastStoreMessage)>kademlia.timeoutFiles{
					if kademlia.data[title].PinStatus{
						fileContact := NewContact(NewHashKademliaId(title),"")
						closestNodes := kademlia.LookupContact(&fileContact)
						Inclosest:= false
						for node:= range closestNodes{
							if kademlia.routingTable.me.ID.String()==closestNodes[node].ID.String(){
								Inclosest=true
							}
						}
						if Inclosest{
							file:=kademlia.data[title]
							go kademlia.Store(&file)
						}else{
							delete(kademlia.data,title)
						}
					}else{
						delete(kademlia.data,title)
					}
				}
			}
		}
		kademlia.mutexData.Unlock()				
		time.Sleep(kademlia.timeoutFiles)
	}
}

func (kademlia *Kademlia) Delete(title string){
	kademlia.mutexFileSend.Lock()
	_, exist := kademlia.filesend[title]
	if exist{
		kademlia.filesend[title].on=false
	}else{
		fmt.Println("File doesn't exist")
	}
	kademlia.mutexFileSend.Unlock()
}

func (kademlia *Kademlia) Pin(title string){
	kademlia.mutexFileSend.Lock()
	_, exist := kademlia.filesend[title]
	if exist{
		kademlia.filesend[title].PinStatus=true
	}else{
		fmt.Println("File doesn't exist")
	}
	kademlia.mutexFileSend.Unlock()
}

func (kademlia *Kademlia) UnPin(title string){
	kademlia.mutexFileSend.Lock()
	_, exist := kademlia.filesend[title]
	if exist{
		kademlia.filesend[title].PinStatus=false
	}else{
		fmt.Println("File doesn't exist")
	}
	kademlia.mutexFileSend.Unlock()
}

func (kademlia *Kademlia) ReceiveMessageUdp(port string) {
	ipAdress := ":" + port
	ServerAddr, err := net.ResolveUDPAddr("udp", ipAdress)
	CheckError(err)
	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	defer ServerConn.Close()

	buf := make([]byte, 1024*4)
	fmt.Println("server ready")
	for {
		//fmt.Println("new loop")
		n, _, err := ServerConn.ReadFromUDP(buf)
		CheckError(err)

		message := string(buf[0:n])
		var decodedMessage Message
		json.Unmarshal([]byte(message), &decodedMessage)

		var responseMessage Message
		responseTcp :=false	
		noResponseNeed :=false
		switch(decodedMessage.MessageType){
			case PING :
				//fmt.Println("Message Ping Received from:", decodedMessage.Source.String())
				kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)				
				responseMessage = Message{decodedMessage.MessageID,kademlia.routingTable.me, RESPONSE , ""}
			break

		case FINDCONTACT:
			//fmt.Println("Message findContact Received:", string(decodedMessage.Content[0]))
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			closestContact := kademlia.routingTable.FindClosestContacts(NewKademliaID(decodedMessage.Content), kademlia.k)
			JSONClosestContact, _ := json.Marshal(closestContact)
			responseMessage = Message{decodedMessage.MessageID, kademlia.routingTable.me, RESPONSE, string(JSONClosestContact)}
			break

		case FINDDATA:
			//fmt.Println("Message findData Received:", string(decodedMessage.Content[0]))
			kademlia.mutexData.Lock()					
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			fileFind, exist := kademlia.data[decodedMessage.Content]
			kademlia.mutexData.Unlock()					
			if exist {
				JSONData, _ := json.Marshal(fileFind)
				responseMessage = Message{decodedMessage.MessageID, kademlia.routingTable.me, DATAFOUND, string(JSONData)}
				responseTcp=true
			} else {
				kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
				closestContact := kademlia.routingTable.FindClosestContacts(NewKademliaID(decodedMessage.Content), kademlia.k)
				JSONData, _ := json.Marshal(closestContact)
				responseMessage = Message{decodedMessage.MessageID, kademlia.routingTable.me, RESPONSE, string(JSONData)}
			}
			break

		case ADDNODE:
			//fmt.Println("Message addNode Received:", string(decodedMessage.Content[0]))
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			closestContact := kademlia.LookupContact(&decodedMessage.Source)
			JSONClosestContact, _ := json.Marshal(closestContact)
			responseMessage = Message{decodedMessage.MessageID, kademlia.routingTable.me, RESPONSE, string(JSONClosestContact)}
			break

		case RESPONSE:
			//fmt.Println("Message RESPONSE Received from:", decodedMessage.Source.String())
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			var contacts []Contact
			json.Unmarshal([]byte(decodedMessage.Content), &contacts)
			var originalMessage Message
			kademlia.network.mutex.Lock()
			originalMessage = kademlia.network.messageMap[decodedMessage.MessageID]
			kademlia.network.mutex.Unlock()
			switch originalMessage.MessageType {
			case FINDCONTACT, FINDDATA:			
				kademlia.routingTable.AddContact(originalMessage.Source, &kademlia.network)
				//fmt.Println("originalMessage")
				//fmt.Println(originalMessage.Source.ID.String())
				kademlia.mutexSem.Lock()
				kademlia.alreadyLookup[decodedMessage.Source.ID.String()] = 2
				kademlia.mutexSem.Unlock()
				kademlia.addToMap(&kademlia.closestContact, &kademlia.alreadyLookup, contacts, NewKademliaID(originalMessage.Content))
				kademlia.semaphore++
				kademlia.nbResponse++			
				break

			default:

				break
			}
			noResponseNeed = true
			break

		default:
			/*fmt.Print("Unexpected Message Received from: ")
			fmt.Println(decodedMessage)*/
			noResponseNeed = true
			break
		}
		if !noResponseNeed {
			//fmt.Println("Response send to ")
			//fmt.Println(decodedMessage.Source)
			if responseTcp {
				kademlia.network.SendMessageTcp(kademlia.routingTable.me, decodedMessage.Source, &responseMessage)										
			}else{
				kademlia.network.SendMessageUdp(kademlia.routingTable.me, decodedMessage.Source, &responseMessage)						
			}
		}
	}
}

func (kademlia *Kademlia) ReceiveMessageTcp(udp_port string) {
	port, _ := strconv.Atoi(udp_port)
	tcp_port := strconv.Itoa(port + 2000)

	l, err := net.Listen("tcp", "localhost:"+tcp_port)
    CheckError(err)
    // Close the listener when the application closes.
    defer l.Close()
    fmt.Println("server tcp ready on port : " + tcp_port)
    for {
        // Listen for an incoming connection.
		conn, err := l.Accept()
        CheckError(err)
        // Handle connections in a new goroutine.
		message, _ := bufio.NewReader(conn).ReadString('\n')
		// output message received
		//fmt.Print("Message Received:", string(message))
		var decodedMessage Message
		
		json.Unmarshal([]byte(message),&decodedMessage)

		var responseMessage Message
		var noResponseNeed bool
		noResponseNeed =false
		switch(decodedMessage.MessageType){

		case STORE:			
			//fmt.Println("Message store tcp Received:", string(decodedMessage.Content[0]))
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			var dataDecoded File
			json.Unmarshal([]byte(decodedMessage.Content), &dataDecoded)
			kademlia.mutexData.Lock()
			kademlia.data[NewHashKademliaId(dataDecoded.Title).String()] = dataDecoded
			kademlia.mutexData.Unlock()			
			responseMessage = Message{decodedMessage.MessageID, kademlia.routingTable.me, RESPONSE, ""}
		break

		case DATAFOUND:
			//fmt.Println("Message RESPONSE Received from:", decodedMessage.Source.String())
			kademlia.routingTable.AddContact(decodedMessage.Source, &kademlia.network)
			var contacts []Contact
			json.Unmarshal([]byte(decodedMessage.Content), &contacts)
			var originalMessage Message
			kademlia.network.mutex.Lock()
			originalMessage = kademlia.network.messageMap[decodedMessage.MessageID]
			kademlia.network.mutex.Unlock()
			if (originalMessage.MessageType==FINDDATA) {
					var file File
					json.Unmarshal([]byte(decodedMessage.Content), &file)
					kademlia.dataFound = file
			}
			noResponseNeed = true
		break

		default:
			/*fmt.Print("Unexpected Message Received from: ")
			fmt.Println(decodedMessage)*/
			noResponseNeed = true
		break
		}
		if !noResponseNeed {
			//fmt.Println("Response send to ")
			//fmt.Println(decodedMessage.Source)
			kademlia.network.SendMessageUdp(kademlia.routingTable.me, decodedMessage.Source, &responseMessage)
		}	
		conn.Close()
	}
}
