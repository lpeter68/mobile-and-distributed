package main

import (
	//"fmt"
	"net"
	//"bufio"
	"encoding/json"
	"strconv"
	//"strings"
	"sync"
	"time"
)

type Network struct {
	mutex      sync.Mutex
	messageLookup map[int]Message
	allMessage map[int]Message	
	indexMap   int 
	kademlia   *Kademlia
	//reqLen []byte
}

type MessageType int

const (
	PING MessageType = 1 + iota
	FINDCONTACT
	FINDDATA
	STORE
	KEEPALIVE
	NEEDFILE
	RESPONSE
	DATAFOUND
)

type Message struct {
	MessageID   int
	Source      Contact
	Destination      Contact	
	MessageType MessageType
	Content     string
	sendingTime time.Time 	//not in JSON
}

type File struct {
	Title string
	Data []byte
	PinStatus bool
	On bool
	LastStoreMessage time.Time 
	changedDetected bool //not in JSON
}

func (network *Network) addMessage(message *Message) {
	network.mutex.Lock()
	if network.messageLookup == nil {
		network.messageLookup = make(map[int]Message)
	}
	if network.allMessage == nil {
		network.indexMap = 0		
		network.allMessage = make(map[int]Message)		
	}
	message.MessageID = network.indexMap
	network.allMessage[network.indexMap] = *message	
	switch message.MessageType{
	case 	FINDCONTACT, FINDDATA :
		network.messageLookup[network.indexMap] = *message			
		break
	}
	network.indexMap = 1 + network.indexMap
	network.mutex.Unlock()
}

func CheckError(err error) {
	if err != nil {
		//fmt.Println("Error: ", err)
	}
}

func (network *Network) SendMessageTcp(sourceContact Contact, contactToSend Contact, message *Message) {
	//Change udp port by tcp port in destination address
	ip, udp_port, _ := net.SplitHostPort(contactToSend.Address)
	port, _ := strconv.Atoi(udp_port)
	tcp_port := strconv.Itoa(port + 2000)

	//New address to the server
	ServerAddr, err2 := net.ResolveTCPAddr("tcp", ip+":"+tcp_port)
	CheckError(err2)

	Conn, err := net.DialTCP("tcp", nil, ServerAddr)
	CheckError(err)
	i := 0
	for err != nil && i<10 {
		CheckError(err)		
		i++
		time.Sleep(300 * time.Millisecond)
		Conn, err = net.DialTCP("tcp", nil, ServerAddr)
	}
	if(err == nil){
		defer Conn.Close()
		text, _ := json.Marshal(message)
		_, err = Conn.Write([]byte(text))
		CheckError(err)
		Conn.Close()
	}
}

func (network *Network) SendMessageUdp(sourceContact Contact, destinationContact Contact, message *Message) {
	ServerAddr, err := net.ResolveUDPAddr("udp", destinationContact.Address)
	CheckError(err)

	Conn, err := net.DialUDP("udp", nil, ServerAddr) 
	CheckError(err)
	i := 0
	for err != nil {
		i++
		time.Sleep(30 * time.Millisecond)
		Conn, err = net.DialUDP("udp", nil, ServerAddr)
		if i > 10 {
			//fmt.Println("infinite loop")
		}
	}

	defer Conn.Close()
	text, _ := json.Marshal(message)
	// send to socket
	_, err = Conn.Write([]byte(string(text) + "\n"))
	CheckError(err)
	//Conn.Close()
}

func (network *Network) SendPingMessage(sourceContact Contact, contactToPing Contact) {
	messageToSend := &Message{0, sourceContact, contactToPing, PING, contactToPing.ID.String(),time.Now()}
	//fmt.Println(messageToSend)
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToPing, messageToSend)
}

func (network *Network) SendFindContactMessage(sourceContact Contact, contactToSend Contact, contactToFind Contact) {
	messageToSend := &Message{0, sourceContact, contactToSend, FINDCONTACT, contactToFind.ID.String(),time.Now()}
	//fmt.Println(messageToSend)
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToSend, messageToSend)
	//fmt.Println("Message is send")
}

func (network *Network) SendFindDataMessage(sourceContact Contact, contactToSend Contact, dataTitle string) {
	dataToFind := NewHashKademliaId(dataTitle)
	messageToSend := &Message{0, sourceContact,contactToSend, FINDDATA, dataToFind.String(),time.Now()}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToSend, messageToSend)
}

func (network *Network) SendStoreMessage(sourceContact Contact, contactToReach Contact, data *File) {
	JSONData, _ := json.Marshal(data)
	messageToSend := &Message{0, sourceContact,contactToReach, STORE, string(JSONData),time.Now()}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )
	//network.addMessage(messageToSend)
	network.SendMessageTcp(sourceContact, contactToReach, messageToSend)
}

func (network *Network) SendKeepAliveMessage(sourceContact Contact, contactToReach Contact, data *File) {
	fileWithoutData := File{data.Title, nil ,data.PinStatus, data.On, data.LastStoreMessage,false}
	JSONData, _ := json.Marshal(fileWithoutData)
	messageToSend := &Message{0, sourceContact,contactToReach, KEEPALIVE, string(JSONData),time.Now()}
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToReach, messageToSend)
}

func (message *Message) String() string {
	return "MessageID : " + strconv.Itoa(message.MessageID) + " Source : " + message.Source.ID.String()  + " Destination : " + message.Destination.ID.String() +" Type : " + string(message.MessageType) + " content : " + message.Content
}
