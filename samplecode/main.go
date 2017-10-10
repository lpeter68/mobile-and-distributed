package main

import (
	"bufio"
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"strconv"
) // Package implementing formatted I/O.

//import "time"

type node struct {
	ip        string
	port      int
	networkID int
}

var k int = 20
var alpha int = 3

func main() {
	var cmd string
	var continueB bool
	continueB = true
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Choose hypervisor mode or standard mode('H'/'S')")
	for continueB{
		cmd, _ = reader.ReadString('\n')
		if len(cmd) > 2 {
			cmd = cmd[:len(cmd)-2]
		}
		part := strings.Split(cmd, " ")	

		switch part[0] {
		case "H" :
			HypervisorMode(reader)
			continueB=false
			break

		case "S" :
			StandardMode(reader)
			continueB=false			
			break

		default :
			if(part[0]!=""){
				fmt.Println("Wrong only 'H' or 'S' are permited")				
			}
			break
		}
	}
	
}

func HypervisorMode(reader *bufio.Reader) {
	var cmd string
	mapKademlia := make(map[string]*Kademlia)
	var continueB bool
	continueB = true
	for continueB == true {
		//time.Sleep(100 * time.Millisecond)

		cmd, _ = reader.ReadString('\n')
		//fmt.Scanln(&cmd)
		//fmt.Println(cmd)
		if len(cmd) > 2 {
			cmd = cmd[:len(cmd)-2]
		}
		part := strings.Split(cmd, " ")

		//fmt.Println(len(part))
		switch part[0] {
		//Change topology region
		case "New":
			rt := NewRoutingTable(NewContact(NewKademliaID(part[1]), part[2]))
			kademlia := NewKademlia(*rt, 20, 3)	
			if len(part) >= 4 {
				mapKademlia[part[3]] = kademlia
			} else {
				mapKademlia[part[1]] = kademlia
			}
			fmt.Println("node "+part[1]+" create")			
			break

		case "Join" :
				mapKademlia[part[1]].JoinNetwork(mapKademlia[part[2]].routingTable.me)
				fmt.Println(part[1]+" join the network")
			break

		case "Refresh" :
			for node := range mapKademlia{
				mapKademlia[node].PingAllContact()				
			}
		break
		
		case "Stop" :
			mapKademlia[part[1]].nodeOn=false;
			delete(mapKademlia,part[1])
			fmt.Println("node "+part[1]+" stoped")						
			break	
		
		
		case "Link":
			mapKademlia[part[1]].PingContact(&mapKademlia[part[2]].routingTable.me)
			mapKademlia[part[2]].PingContact(&mapKademlia[part[1]].routingTable.me)
			fmt.Println("link establish beetween "+part[1]+" and"+part[2])						
			break
		
		//Store File region
		case "Store": //TODO
			fmt.Println("Id for data : " + NewHashKademliaId(part[2]).String())
			path := strings.Split(part[2], "/")
			path = strings.Split(path[len(path)-1], "\\")
			fileName := path[len(path)-1]
			fmt.Println(fileName)
			file, err := os.Open(part[2])
			CheckError(err)
			stat,_ := file.Stat()
			size := stat.Size()
			var fileBuffer []byte = make([]byte, size)
			for {
				_, err := file.Read(fileBuffer)
				if err == io.EOF {
					break
				}
				go mapKademlia[part[1]].Store(&File{fileName, fileBuffer,false,true,time.Now(),false})
			}
			fmt.Println("Data stored")			
			break

		case "Pin" :
			go mapKademlia[part[1]].PinFile(part[2])
			fmt.Println("file pin")			
			break

		case "UnPin" : 
			go mapKademlia[part[1]].UnPinFile(part[2])
			fmt.Println("file unPin")						
			break

		case "Delete" : 
			go mapKademlia[part[1]].DeleteFile(part[2])
			fmt.Println("file delete (it will be efficient in some time)")									
			break
		
		//Find region
		case "FindData":
			fmt.Println("FindData command on node ")
			fmt.Println(part[1])
			fmt.Println(string(mapKademlia[part[1]].LookupData(part[2])))
			break

		case "FindNode":
			fmt.Println("FindNode command on node ")
			fmt.Println(part[1])
			_, exist := mapKademlia[part[2]]
			var result []Contact
			if exist {
				result = mapKademlia[part[1]].LookupContact(&mapKademlia[part[2]].routingTable.me)
			} else {
				tar := NewContact(NewKademliaID(part[2]), "")
				result = mapKademlia[part[1]].LookupContact(&tar)
			}
			fmt.Println(" k closest contact find from " + part[1])
			for i := range result {
				fmt.Println("ID : " + giveName(mapKademlia, result[i].ID.String()) + " dist : " + result[i].distance.String())
			}
			break

		
		//tools region
		case "PrintMap":
			/*fmt.Println("map lenght "+string(len(mapKademlia)))
			for i := range mapKademlia {
				fmt.Println("ID "+i+" : "+ mapKademlia[i].routingTable.me.ID.String())
			}*/
			graphvizContent := "graph {\n "
			if len(part) > 2 {
				graphvizContent += mapKademlia[part[2]].routingTable.me.ID.String() + "[label=" + part[2]
				if len(mapKademlia[part[2]].data) > 0 {
					graphvizContent += " style=filled fillcolor=yellow"
				}
				graphvizContent += "];\n"
				for j := range mapKademlia[part[2]].routingTable.buckets {
					for k := mapKademlia[part[2]].routingTable.buckets[j].list.Front(); k != nil; k = k.Next() {
						nodeIDPseudo := k.Value.(Contact).ID
						nodeID := giveName(mapKademlia, nodeIDPseudo.String())
						reverseLink := nodeID + " -- " + mapKademlia[part[2]].routingTable.me.ID.String() + ";"
						if !strings.Contains(graphvizContent, reverseLink) {
							graphvizContent += mapKademlia[part[2]].routingTable.me.ID.String() + " -- " + nodeID + ";\n"
						}
					}
				}
			} else {
				for i := range mapKademlia {
					graphvizContent += mapKademlia[i].routingTable.me.ID.String() + "[label=" + i
					if len(mapKademlia[i].data) > 0 {
						graphvizContent += " style=filled fillcolor=yellow"
					}
					graphvizContent += "];\n"
					for j := range mapKademlia[i].routingTable.buckets {
						for k := mapKademlia[i].routingTable.buckets[j].list.Front(); k != nil; k = k.Next() {
							nodeID := k.Value.(Contact).ID
							reverseLink := nodeID.String() + " -- " + mapKademlia[i].routingTable.me.ID.String() + ";"
							if !strings.Contains(graphvizContent, reverseLink) {
								graphvizContent += mapKademlia[i].routingTable.me.ID.String() + " -- " + nodeID.String() + ";\n"
							}
						}
					}
				}
			}
			graphvizContent += "}"
			ioutil.WriteFile(part[1], []byte(graphvizContent), 0644)
			fmt.Println("graphe generate to .dot format")									
			break

		case "help":
			fmt.Println("Command available are :")
			fmt.Println("	New <KademliaID> <ip:port> <pseudo>")
			fmt.Println("		Create a new node")
			fmt.Println("")
			fmt.Println("	Join <pseudo> <ip:port>")
			fmt.Println("		Add the node called pseudo to the network ip:port must be a node on the network")
			fmt.Println("")
			fmt.Println("	Stop <pseudo> ")
			fmt.Println("		Stop the node called pseudo ")
			fmt.Println("")
			fmt.Println("	Link <pseudo1> <pseudo 2>")
			fmt.Println("		Add pseudo1 to the pseudo2 routing table and vise versa")
			fmt.Println("")
			fmt.Println("	Store <pseudo1> <file>")
			fmt.Println("		pseudo1 store the file on the network")
			fmt.Println("")
			fmt.Println("	Pin/UnPin <pseudo1> <file>")				
			fmt.Println("		pseudo1 pin or unpin a file already stored")
			fmt.Println("")
			fmt.Println("	Delete <pseudo1> <file>")				
			fmt.Println("		pseudo1 delete a file stored")
			fmt.Println("")	
			fmt.Println("	FindData <pseudo1> <file> ")
			fmt.Println("		pseudo1 shearch a file on the netwotk ")
			fmt.Println("")
			fmt.Println("	FindNode <pseudo1> <KademliaID> ")
			fmt.Println("		pseudo1 shearch closest node from KademliaID on network ")
			fmt.Println("")
			fmt.Println("	PrintMap <outputFile> optionnal <pseudo>")
			fmt.Println("		generate a graph to .dot format of pseudo routing table or all the network")
			fmt.Println("")
			fmt.Println("	Exit")
			fmt.Println("		end of the simulation")
			break

		
		case "Exit":
			continueB = false
			break
		}
	}
}

func giveName(mapKademlia map[string]*Kademlia, objective string) string {
	for i := range mapKademlia {
		if strings.ToUpper(mapKademlia[i].routingTable.me.ID.String()) == strings.ToUpper(objective) {
			return i
		}
	}
	return ""
}

func StandardMode(reader *bufio.Reader) {
	var cmd string
	var node *Kademlia
	nodeInit := false
	continueB := true
	fmt.Println("Enter port number to use : Warning the port +2000 is also use")
	for continueB == true {
		time.Sleep(1000 * time.Millisecond)

		cmd, _ = reader.ReadString('\n')
		//fmt.Scanln(&cmd)
		//fmt.Println(cmd)
		if len(cmd) > 2 {
			cmd = cmd[:len(cmd)-2]
		}
		part := strings.Split(cmd, " ")

		if(!nodeInit && part[0]!=""){
			port,_ := strconv.Atoi(part[0])
			if(port>1024){
				rt := NewRoutingTable(NewContact(NewHashKademliaId(part[0]), "localhost:"+part[0]))
				node = NewKademlia(*rt, 20, 3)	
				fmt.Println("your id is : "+node.routingTable.me.ID.String())
				nodeInit=true
			}else{
				println(part[0])
				println("Invalid port")
			}
		}else{
			//fmt.Println(len(part))
			switch part[0] {

			case "Join" :
					node.JoinNetwork(NewContact(NewHashKademliaId(part[1]), "localhost:"+part[1]))
					fmt.Println("you join the network")					
				break
			
			//Store File region
			case "Store": //TODO
				fmt.Println("Id for data : " + NewHashKademliaId(part[1]).String())
				path := strings.Split(part[1], "/")
				path = strings.Split(path[len(path)-1], "\\")				
				fileName := path[len(path)-1]
				fmt.Println(fileName)
				file, err := os.Open(fileName)
				CheckError(err)
				stat,_ := file.Stat()
				size := stat.Size()
				var fileBuffer []byte = make([]byte, size)
				for {
					_, err := file.Read(fileBuffer)
					if err == io.EOF {
						break
					}
					go node.Store(&File{fileName, fileBuffer,false,true,time.Now(),false})
				}
				fmt.Println("file stored")										
				break

			case "Pin" :
				go node.PinFile(part[1])
				fmt.Println("file pin")										
				break

			case "UnPin" : 
				go node.UnPinFile(part[1])
				fmt.Println("file unPin")										
				break

			case "Delete" : 
				go node.DeleteFile(part[1])
				fmt.Println("file delete (it will be efficient in some time)")									
				break
			
			//Find region
			case "FindData":
				fmt.Println(string(node.LookupData(part[1])))
				break

			case "FindNode":
				var result []Contact
				tar := NewContact(NewKademliaID(part[1]), "")
				result = node.LookupContact(&tar)
				fmt.Println(" k closest contact find from " + part[1])
				for i := range result {
					fmt.Println(result[i])
				}
				break

			
			//tools region
			case "PrintMap":
				/*fmt.Println("map lenght "+string(len(mapKademlia)))
				for i := range mapKademlia {
					fmt.Println("ID "+i+" : "+ mapKademlia[i].routingTable.me.ID.String())
				}*/
				graphvizContent := "graph {\n "
				graphvizContent += "\""+node.routingTable.me.ID.String() + "\"[label=\"" + node.routingTable.me.ID.String() +"\""
				if len(node.data) > 0 {
					graphvizContent += " style=filled fillcolor=yellow"
				}
				graphvizContent += "];\n"
				for j := range node.routingTable.buckets {
					for k := node.routingTable.buckets[j].list.Front(); k != nil; k = k.Next() {
						nodeID := k.Value.(Contact).ID.String()
						reverseLink := nodeID + " -- " + node.routingTable.me.ID.String() + ";"
						if !strings.Contains(graphvizContent, reverseLink) {
							graphvizContent += "\""+node.routingTable.me.ID.String() + "\" -- \"" + nodeID + "\";\n"
						}
					}
				}			
				graphvizContent += "}"
				ioutil.WriteFile(part[1], []byte(graphvizContent), 0644)
				fmt.Println("graph generate to .dot format")													
				break

			case "help":
				fmt.Println("	Join <port>")
				fmt.Println("		Add the node to the network, port must be used by a node on the network")
				fmt.Println("")
				fmt.Println("	Store <file>")
				fmt.Println("		store the file on the network")
				fmt.Println("")
				fmt.Println("	Pin/UnPin <file>")				
				fmt.Println("		pin or unpin a file already stored")
				fmt.Println("")
				fmt.Println("	Delete <file>")				
				fmt.Println("		delete a file stored")
				fmt.Println("")	
				fmt.Println("	FindData <file> ")
				fmt.Println("		shearch a file on the netwotk ")
				fmt.Println("")
				fmt.Println("	FindNode <KademliaID> ")
				fmt.Println("		shearch closest node from KademliaID on network ")
				fmt.Println("")
				fmt.Println("	PrintMap <outputFile>")
				fmt.Println("		generate a graph to .dot format of our routing table ")
				fmt.Println("")
				fmt.Println("	Exit")
				fmt.Println("		Close the app")
				break

			
			case "Exit":
				node.nodeOn=false;			
				continueB = false
				break
			}
		}
	}
}

