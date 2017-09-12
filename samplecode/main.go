package main

import fmt "fmt" // Package implementing formatted I/O.
import "strings"
import "bufio"
import "os"
import "io/ioutil"

type node struct {
	ip string
	port int
	networkID int
}

var k int = 20
var alpha int =3

/*func main() {

	/* rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))	 
	//rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8007"))
	//rt.AddContact(NewContact(NewKademliaID("1111111210000000000000000000000000000000"), "localhost:8004"))
	
	
	kademlia := NewKademlia(*rt,2,1)
	 go kademlia.ReceiveMessage("8000")

	 rt2 := NewRoutingTable(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8007"))	 
	 
	 kademlia2 := NewKademlia(*rt2,2,1)
	  go kademlia2.ReceiveMessage("8007")

	 var contact Contact=NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	 var target Contact=NewContact(NewKademliaID("1111111210000000000000000000000000000000"), "localhost:8008")
	 if(SendPingMessage(target,contact)){
		fmt.Println("Success")		
	 }

	 rt2 := NewRoutingTable(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))	 
	 rt2.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))
	 rt2.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	 rt2.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8003"))
	 rt2.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8004"))
	 rt2.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8005"))
	 rt2.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8006"))
 
	 kademlia2 := NewKademlia(*rt2,2,1)
	  go kademlia2.ReceiveMessage("8002")

	  rt3 := NewRoutingTable(NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8003"))	 
	  rt3.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))
	  //rt3.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	  rt3.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8003"))
	  rt3.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8004"))
	  rt3.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8005"))
	  rt3.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8006"))
  
	  kademlia3 := NewKademlia(*rt3,2,1)
	   go kademlia3.ReceiveMessage("8003")

	 var contact Contact=NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	 var contact2 Contact=NewContact(NewKademliaID("FFFFFFF100000000000000000000000000000000"), "localhost:8001")	 
	 var target Contact=NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8003")	 
	 SendStoreMessage(contact,target,File{"great",[]byte("test result")})
	 fmt.Println(string(SendFindDataMessage(target,"great",contact2)))
	 
}*/

func main() {
	var cmd string
	mapKademlia :=  make(map[string]*Kademlia)
	reader := bufio.NewReader(os.Stdin)
	for ;;{
		cmd, _ = reader.ReadString('\n')
		//fmt.Scanln(&cmd)
		//fmt.Println(cmd)
		if(len(cmd)>2){
			cmd = cmd[:len(cmd)-2]				
		}
		part := strings.Split(cmd," ")
		
		//fmt.Println(len(part))
		switch(part[0]){
			case "New" :
				rt := NewRoutingTable(NewContact(NewKademliaID(part[1]), part[2]))
				kademlia := NewKademlia(*rt,2,1)			
				go kademlia.ReceiveMessage(strings.Split(part[2],":")[1])
				if(len(part)>=4){
					mapKademlia[part[3]]=kademlia
				}else{
					mapKademlia[part[1]]=kademlia					
				}
			break

			case "Join" :
				mapKademlia[part[1]].AddToNetwork(part[2])
			break

			case "Link" :									
				mapKademlia[part[1]].PingContact(&mapKademlia[part[2]].routingTable.me)
				mapKademlia[part[2]].PingContact(&mapKademlia[part[1]].routingTable.me)
			break
			
			case "Store" :	
				fmt.Println("Id for data : "+NewHashKademliaId(part[2]).String())			
				mapKademlia[part[1]].Store( File{part[2],[]byte(part[3])})
			break

			case "FindData" :
				fmt.Println(string(mapKademlia[part[1]].LookupData(part[2])))
			break

			case "FindNode" :
				_, exist :=mapKademlia[part[2]]
				if(exist){
					mapKademlia[part[1]].LookupContact(&mapKademlia[part[2]].routingTable.me)
				}else{
					tar :=NewContact(NewKademliaID(part[2]), "")
					mapKademlia[part[1]].LookupContact(&tar)
				}
			break

			case "PrintMap" :
				/*fmt.Println("map lenght "+string(len(mapKademlia)))
				for i := range mapKademlia {
					fmt.Println("ID "+i+" : "+ mapKademlia[i].routingTable.me.ID.String())	
				}*/
				graphvizContent := "graph {\n "
				for i:= range mapKademlia {
					graphvizContent+=mapKademlia[i].routingTable.me.ID.String() +"[label="+i
					if(len(mapKademlia[i].data)>0){
						graphvizContent+=" style=filled fillcolor=yellow"						
					}
					graphvizContent+="];\n"	
					for j:= range mapKademlia[i].routingTable.buckets{
						for k := mapKademlia[i].routingTable.buckets[j].list.Front(); k != nil; k = k.Next() {
							nodeID := k.Value.(Contact).ID
							reverseLink := nodeID.String()+" -- "+mapKademlia[i].routingTable.me.ID.String() +";"
							if(!strings.Contains(graphvizContent,reverseLink)){
								graphvizContent+=mapKademlia[i].routingTable.me.ID.String()+" -- "+nodeID.String() +";\n"								
							}
						}
					}
				}
				graphvizContent+="}"
				ioutil.WriteFile(part[1], []byte(graphvizContent), 0644)
			break
		}
	}
}