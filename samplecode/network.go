package main

import (
	"fmt"
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
	messageMap map[int]Message
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
	ADDNODE
	RESPONSE
	DATAFOUND
)

type Message struct {
	MessageID   int
	Source      Contact
	MessageType MessageType
	Content     string
}

type File struct {
	Title string
	Data  []byte
}

func (network *Network) ListenTcp(sourceContact Contact, contactToSend Contact) {
	//Change udp port by tcp port in destination address
	ip, udp_port, _ := net.SplitHostPort(contactToSend.Address)
	port, _ := strconv.Atoi(udp_port)
	tcp_port := strconv.Itoa(port + 2000)
	//New address to the server
	ServerAddr, err := net.ResolveTCPAddr("tcp", ip+":"+tcp_port)
	CheckError(err)
	//Change udp port by tcp port in source address
	ip, udp_port, _ = net.SplitHostPort(sourceContact.Address)
	port, _ = strconv.Atoi(udp_port)
	tcp_port = strconv.Itoa(port + 2000)
	//New address to the source
	LocalAddr, err := net.ResolveTCPAddr("tcp", ip+":"+tcp_port)
	CheckError(err)
	//Stablishing connection
	Conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	CheckError(err)
	network.kademlia.storeInMap(Conn)

	Conn.Close()

}

func (kademlia *Kademlia) storeInMap(Conn net.Conn) {
	buf := make([]byte, 1024)
	var title string
	kademlia.mutexData.Lock()
	// Read the name of the file.
	Conn.Read(buf)
	title = string(buf)
	//fileName = Conn.Read([]byte(string))
	//var dataDecoded Message
	//json.Unmarshal([]byte(buf), &dataDecoded)
	//kademlia.data[title] = Conn.Read(title)

	//kademlia.data[fileName].Title = fileName
	Conn.Read(kademlia.data[title].Data)
	kademlia.mutexData.Unlock()

}

func (network *Network) addMessage(message *Message) {
	network.mutex.Lock()
	if network.messageMap == nil {
		network.indexMap = 0
		network.messageMap = make(map[int]Message)
	}
	message.MessageID = network.indexMap
	//fmt.Println("message envoyé et stocké")
	//fmt.Println(message)
	network.messageMap[network.indexMap] = *message
	network.indexMap = 1 + network.indexMap
	network.mutex.Unlock()
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (network *Network) SendMessageTcp(sourceContact Contact, contactToSend Contact, file *File) {
	//Change udp port by tcp port in destination address
	ip, udp_port, _ := net.SplitHostPort(contactToSend.Address)
	port, _ := strconv.Atoi(udp_port)
	tcp_port := strconv.Itoa(port + 2000)
	//New address to the server
	ServerAddr, err := net.ResolveTCPAddr("tcp", ip+":"+tcp_port)
	CheckError(err)
	//Change udp port by tcp port in source address
	ip, udp_port, _ = net.SplitHostPort(sourceContact.Address)
	port, _ = strconv.Atoi(udp_port)
	tcp_port = strconv.Itoa(port + 2000)
	//New address to the source
	LocalAddr, err := net.ResolveTCPAddr("tcp", ip+":"+tcp_port)
	CheckError(err)
	//Stablishing connection
	Conn, err := net.DialTCP("tcp", LocalAddr, ServerAddr)
	CheckError(err)
	i := 0
	for err != nil {
		i++
		time.Sleep(30 * time.Millisecond)
		Conn, err = net.DialTCP("tcp", LocalAddr, ServerAddr)
		if i > 10 {
			//fmt.Println("infinite loop")
		}
	}

	defer Conn.Close()
	//text, _ := json.Marshal(message)
	// first is sended the title of the file
	_, err = Conn.Write([]byte(file.Title))
	//send the content of the file
	_, err = Conn.Write(file.Data)
	CheckError(err)
	//Conn.Close()
}

func (network *Network) SendMessageUdp(sourceContact Contact, destinationContact Contact, message *Message) {
	ServerAddr, err := net.ResolveUDPAddr("udp", destinationContact.Address)
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", sourceContact.Address)
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)
	i := 0
	for err != nil {
		i++
		time.Sleep(30 * time.Millisecond)
		Conn, err = net.DialUDP("udp", LocalAddr, ServerAddr)
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
	messageToSend := &Message{0, sourceContact, PING, contactToPing.ID.String()}
	//fmt.Println(messageToSend)
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToPing, messageToSend)
}

func (network *Network) SendFindContactMessage(sourceContact Contact, contactToSend Contact, contactToFind Contact) {
	messageToSend := &Message{0, sourceContact, FINDCONTACT, contactToFind.ID.String()}
	//fmt.Println(messageToSend)
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToSend, messageToSend)
	//fmt.Println("Message is send")
}

func (network *Network) SendFindDataMessage(sourceContact Contact, contactToSend Contact, dataTitle string) {
	dataToFind := NewHashKademliaId(dataTitle)
	messageToSend := &Message{0, sourceContact, FINDDATA, dataToFind.String()}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToSend, messageToSend)
}

func (network *Network) SendStoreMessage(sourceContact Contact, contactToReach Contact, data File) {
	JSONData, _ := json.Marshal(data)
	messageToSend := &Message{0, sourceContact, STORE, string(JSONData)}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )
	network.addMessage(messageToSend)
	network.SendMessageUdp(sourceContact, contactToReach, messageToSend)
}

func (message *Message) String() string {
	return "MessageID : " + strconv.Itoa(message.MessageID) + " Source : " + message.Source.ID.String() + " Type : " + string(message.MessageType) + " content : " + message.Content
}
