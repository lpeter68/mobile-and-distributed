package main

import (
    "fmt"
    "net"
    "strconv"
    "bufio"
    //"os"
    "encoding/json"
    "time"
)


type Network struct {
}

type MessageType int

const (
   PING MessageType = 1 + iota
   FINDCONTACT
   FINDDATA
   STORE
   ADDNODE
   RESPONSE
)

type Message struct {
	Source Contact
	MessageType MessageType
	Content string
}

type File struct {
	Title string
	Data []byte
}

func Listen(ip string, port int) {
  // TODO
    newport := strconv.Itoa(port)
/* Lets prepare an address at any address at port given*/
    ServerAddr,err := net.ResolveUDPAddr("udp",ip + ":" + newport)
    /* Now listen at selected port */
    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    defer ServerConn.Close()
  /*I dont really know from this to the end of the function*/
  buf := make([]byte, 1024)

  for {
      n,addr,err := ServerConn.ReadFromUDP(buf)
      fmt.Println("Received ",string(buf[0:n]), " from ",addr)
  }
}

func SendPingMessage( sourceContact Contact, contactToPing Contact ) bool {

	  // listen for reply
	  input := make(chan string, 1) /*Create a channel*/
	  go getInput(input, contactToPing, sourceContact)

	for {
		select {
		case i := <-input:
			var message Message
			json.Unmarshal([]byte(i),&message)
			if(message.MessageType==RESPONSE){
				return true
			}
		case <-time.After(4000 * time.Millisecond):
			fmt.Println("timed out")
			return false
		}
	}
}

func getInput(input chan string, contactToPing Contact, sourceContact Contact) {
    for {
		messageToSend := &Message{sourceContact, PING,contactToPing.ID.String()}
		//fmt.Println("messageToSend to messageToSend server: "+messageToSend.Content )

		conn, conErr := net.Dial("tcp", contactToPing.Address)
		//fmt.Println(conErr)
		if(conErr==nil){
			//fmt.Println("Text to send: ")
			text, err := json.Marshal(messageToSend)
			if (err != nil) {
				fmt.Println("error " )
				fmt.Println(err)
			}
			//fmt.Println("Message to send server: "+string(text))

			// send to socket
			fmt.Fprintf(conn, string(text) + "\n")
			JSONmessage, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
			}
			input <- JSONmessage
		}
    }
}

func SendFindContactMessage(sourceContact Contact, contactToSend Contact, contactToFind Contact) []Contact {
	messageToSend := &Message{sourceContact, FINDCONTACT,contactToFind.ID.String()}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )

	conn, _ := net.Dial("tcp", contactToSend.Address)
	//	  fmt.Print("Text to send: ")
	  text, err := json.Marshal(messageToSend)
	  if err != nil {
		fmt.Println("error " )
		fmt.Println(err)
	}
	  //fmt.Println("Message to send server: "+string(text))

	  // send to socket
	  fmt.Fprintf(conn, string(text) + "\n")
	  // listen for reply
	  JSONmessage, _ := bufio.NewReader(conn).ReadString('\n')
	  var message Message
	  json.Unmarshal([]byte(JSONmessage),&message)
	  var contacts []Contact
	  json.Unmarshal([]byte(message.Content),&contacts)
	  /*for i := range contacts {
	  	fmt.Println("Message from server " +string(i) +" : "+ contacts[i].ID.String())
	  }*/
	  return contacts
}

func SendFindDataMessage(sourceContact Contact, contactToSend Contact, dataTitle string) []byte {
	dataToFind := NewHashKademliaId(dataTitle)
	messageToSend := &Message{sourceContact, FINDDATA, dataToFind.String()}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )

	conn, _ := net.Dial("tcp", contactToSend.Address)
	//	  fmt.Print("Text to send: ")
	  text, err := json.Marshal(messageToSend)
	  if err != nil {
		fmt.Println("error " )
		fmt.Println(err)
	}
	  //fmt.Println("Message to send server: "+string(text))

	  // send to socket
	  fmt.Fprintf(conn, string(text) + "\n")
	  // listen for reply
	  JSONmessage, _ := bufio.NewReader(conn).ReadString('\n')
	  var message Message
	  json.Unmarshal([]byte(JSONmessage),&message)
	  if(message.Content!=""){
		var file File
		json.Unmarshal([]byte(message.Content),&file)
		return file.Data
	  }else{
		return nil
	  }

	  /*for i := range contacts {
	  	fmt.Println("Message from server " +string(i) +" : "+ contacts[i].ID.String())
	  }*/
}

func SendStoreMessage(sourceContact Contact, contactToReach Contact, data File) {
	JSONData, _ := json.Marshal(data)
	messageToSend := &Message{sourceContact, STORE,string(JSONData)}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )

	conn, _ := net.Dial("tcp", contactToReach.Address)
	//	  fmt.Print("Text to send: ")
	  text, err := json.Marshal(messageToSend)
	  if err != nil {
		fmt.Println("error " )
		fmt.Println(err)
	}
	  //fmt.Println("Message to send server: "+string(text))

	  // send to socket
	  fmt.Fprintf(conn, string(text) + "\n")
	  // listen for reply
	  JSONmessage, _ := bufio.NewReader(conn).ReadString('\n')
	  var message Message
	  json.Unmarshal([]byte(JSONmessage),&message)
}

func SendAddNodeMessage(sourceContact Contact, address string) []Contact {
	messageToSend := &Message{sourceContact, ADDNODE,""}
	//fmt.Print("messageToSend to messageToSend server: "+messageToSend.Content )

	conn, _ := net.Dial("tcp", address)
	//	  fmt.Print("Text to send: ")
	  text, err := json.Marshal(messageToSend)
	  if err != nil {
		fmt.Println("error " )
		fmt.Println(err)
	}
	  //fmt.Println("Message to send server: "+string(text))

	  // send to socket
	  fmt.Fprintf(conn, string(text) + "\n")
	  // listen for reply
	  JSONmessage, _ := bufio.NewReader(conn).ReadString('\n')
	  var message Message
	  json.Unmarshal([]byte(JSONmessage),&message)
	  var contacts []Contact
	  json.Unmarshal([]byte(message.Content),&contacts)
	  return contacts
}

func (message *Message) String() string {
	return "Source : "+message.Source.ID.String()+" Type : "+ string(message.MessageType)+ " content : "+ message.Content
}
