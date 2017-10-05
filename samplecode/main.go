package main

import (
	"bufio"
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
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
	//TestPingNode2()
	var cmd string
	mapKademlia := make(map[string]*Kademlia)
	reader := bufio.NewReader(os.Stdin)
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
		case "New":
			rt := NewRoutingTable(NewContact(NewKademliaID(part[1]), part[2]))
			kademlia := NewKademlia(*rt, 20, 3)
			go kademlia.ReceiveMessage(strings.Split(part[2], ":")[1])
			if len(part) >= 4 {
				mapKademlia[part[3]] = kademlia
			} else {
				mapKademlia[part[1]] = kademlia
			}
			break

			case "Join" :
				mapKademlia[part[1]].JoinNetwork(mapKademlia[part[2]].routingTable.me)
			break

		case "Link":
			mapKademlia[part[1]].PingContact(&mapKademlia[part[2]].routingTable.me)
			mapKademlia[part[2]].PingContact(&mapKademlia[part[1]].routingTable.me)
			break

		case "Store": //TODO
			fmt.Println("Id for data : " + NewHashKademliaId(part[2]).String())
			fileName := part[1]
			file, err := os.Open(part[1])
			CheckError(err)
			fileBuffer := make([]byte, 1024)
			for {
				_, err = file.Read(fileBuffer)
				if err == io.EOF {
					break
				}
				mapKademlia[part[1]].Store(File{fileName, fileBuffer})
			}
			break

		case "FindData":
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
			break

		case "help":
			fmt.Println("Command available are :")
			fmt.Println("	New <KademliaID> <ip:port> <pseudo>")
			fmt.Println("		Create a new node")
			fmt.Println("")
			fmt.Println("	Join <pseudo> <ip:port>")
			fmt.Println("		Add the node called pseudo to the network ip:port must be a node on the network")
			fmt.Println("")
			fmt.Println("	Link <pseudo1> <pseudo 2>")
			fmt.Println("		Add pseudo1 to the pseudo2 routing table and vise versa")
			fmt.Println("")
			fmt.Println("	Store <pseudo1> <title> <content>")
			fmt.Println("		pseudo1 publish a file to store with title and content")
			fmt.Println("")
			fmt.Println("	FindData <pseudo1> <title> ")
			fmt.Println("		pseudo1 shearch file with the name title on the netwotk ")
			fmt.Println("")
			fmt.Println("	FindNode <pseudo1> <KademliaID> ")
			fmt.Println("		pseudo1 shearch closest node from KademliaID on network ")
			fmt.Println("")
			fmt.Println("	PrintMap <outputFile> optionnal <pseudo>")
			fmt.Println("		generate a graph to .dot format of pseudo routing table or all the network")
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
