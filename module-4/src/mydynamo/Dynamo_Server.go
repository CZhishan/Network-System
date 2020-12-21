package mydynamo

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"errors"
	"time"
	"sync"
)

type DynamoServer struct {
	/*------------Dynamo-specific-------------*/
	wValue         int          //Number of nodes to write to on each Put
	rValue         int          //Number of nodes to read from on each Get
	preferenceList []DynamoNode //Ordered list of other Dynamo nodes to perform operations o
	selfNode       DynamoNode   //This node's address and port info
	nodeID         string       //ID of this node
	dataStore      map[string][]ObjectEntry //Data stored in server, with a key associated with
	crashed        bool
	gossipList 	   map[int]PutArgs
}

func (s *DynamoServer) SendPreferenceList(incomingList []DynamoNode, _ *Empty) error {
	log.Println("Preferencelist sent!")
	s.preferenceList = incomingList
	return nil
}

// Forces server to gossip
// As this method takes no arguments, we must use the Empty placeholder
func (s *DynamoServer) Gossip(_ Empty, _ *Empty) error {
	templist := s.gossipList
	var m sync.Mutex
	var err error
	for i, value := range templist {
		node := s.preferenceList[i]

		c, err := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		if err != nil {
			//if the connection failed: proceed to the next node on the list
			log.Println("Failed to connect to the server")
			continue
		}
		var result bool
		err = c.Call("MyDynamo.LocalPut", value,&result)
		if err!=nil{
			//if the server crashed, proceed to the next node
			log.Println("server crashed!")
			continue
		}
		m.Lock()
		delete(s.gossipList, i)
		m.Unlock()
	}
	return err
}

func sleep(seconds int, s *DynamoServer) {
	time.Sleep(time.Duration(seconds) * time.Second)
	s.crashed = false
}

//Makes server unavailable for some seconds
func (s *DynamoServer) Crash(seconds int, success *bool) error {
	s.crashed = true
	go sleep(seconds, s)
	*success = true
	return nil
}

/*func checkCausality(localEntry []ObjectEntry, newEntry ObjectEntry) []ObjectEntry{
	//record the number of entries that is concurrent with the newEntry
	concurrentNum :=0
	for _, obj := range localEntry{
		//if the new Entry is concurrent with node[i], increment the count by 1
		if newEntry.Context.Clock.Concurrent(obj.Context.Clock){
			concurrentNum+=1
		}
	}
	if concurrentNum==len(localEntry){
		//log.Println("value 1:" + string(newEntry.Value[:]))
		//log.Println("localValue"+ string(localEntry[0].Value[:]))
		//if the newEntry is concurrent with all the entries in local,append it
		localEntry = append(localEntry,newEntry)
	}else{
		for i, obj := range localEntry {
			if obj.Context.Clock.LessThan(newEntry.Context.Clock){
				//if the new Entry is calsally descended from the current Entry
				//replace the current with the new
				localEntry[i] = newEntry
			}
		}
	}
	return localEntry
}*/

func checkCausality(localEntry []ObjectEntry, newEntry ObjectEntry) []ObjectEntry {
	var templist []ObjectEntry
	causal := true
	for _, obj := range localEntry {
		if newEntry.Context.Clock.LessThan(obj.Context.Clock){
			causal = false
			templist = append(templist, obj)
		} else if newEntry.Context.Clock.Concurrent(obj.Context.Clock){
			templist = append(templist, obj)
		} else if newEntry.Context.Clock.Equals(obj.Context.Clock){
			causal = false
			templist = append(templist, obj)
		}
	}
	if causal {
		templist = append(templist, newEntry)
	} 
	return templist
}


func (s *DynamoServer) LocalPut(value PutArgs,result *bool) error {
	if s.crashed {
		return errors.New("Sever crashed!")
	}
	key := value.Key
	newObj := ObjectEntry{Context: value.Context, Value: value.Value}

	if _, ok := s.dataStore[key]; ok {
	//if key is already in the datastore
		//if the stored context is causally descended from the context of the input value
		s.dataStore[key] = checkCausality(s.dataStore[key], newObj)
	} else {
		//if no clock conflicts: update the fileds of dataStore
		s.dataStore[key] = append(s.dataStore[key], newObj)
	}
	return nil
}


// Put a file to this server and W other servers
func (s *DynamoServer) Put(value PutArgs, result *bool) error {
	//increment the vectorclock associated with this node
	value.Context.Clock.Increment(s.nodeID)

	var r bool
	err := s.LocalPut(value,&r)
	if err != nil {
		return err
	}

	//replicate the operation for top W nodes in the preferencelist
	writeNum := s.wValue-1
	var replicatedList []int
	for i,node:=range s.preferenceList{
		if s.selfNode.Port==node.Port{
			//skip selfNode
			continue
		}
		if writeNum ==0{
			break
		}
		c, err := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		if err != nil {
			//if the connection failed: proceed to the next node on the list
			log.Println("Failed to connect to the server")
			continue
		}
		err = c.Call("MyDynamo.LocalPut", value,&r)
		if err!=nil{
			//if the server crashed, proceed to the next node
			log.Println("server crashed!")
			continue
		}
		// record the index of the nodes that has replicated the operation
		replicatedList = append(replicatedList, i)
		writeNum-=1
	}
	if writeNum == 0 {
		*result = true
	} else {
		*result = false
	}

	//To do: record lists for nodes that didn't replicate the opeation.
	var m sync.Mutex
	for i, _ := range s.preferenceList {
		if contains(replicatedList, i) {
			continue
		}
		m.Lock()
		s.gossipList[i] = value
		m.Unlock()
	}
	return nil
}


func (s *DynamoServer) LocalGet(key string, result *DynamoResult) error {
	if s.crashed {
		return errors.New("Sever crashed!")
	}

	if localEntry, ok := s.dataStore[key]; ok {
		for _, entry := range localEntry {
			result.EntryList = checkCausality(result.EntryList, entry)
		}
	}
	return nil
}

//Get a file from this server, matched with R other servers
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	if s.crashed {
		return errors.New("Sever crashed!")
	}
	if entry, ok := s.dataStore[key]; ok {
		result.EntryList = append(result.EntryList, entry...)
	}

	readNum := s.rValue-1
	for _,node :=range s.preferenceList{
		if s.selfNode.Port == node.Port{
			continue
		}
		if readNum==0{
			break
		}
		c, err := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		if err!= nil{
			log.Println("Failed to connect to the server")
			continue
		}
		err = c.Call("MyDynamo.LocalGet", key, result)
		if err!=nil{
			log.Println("server crashed!")
			continue
		}
		readNum-=1
	}
	return nil
}

/* Belows are functions that implement server boot up and initialization */
func NewDynamoServer(w int, r int, hostAddr string, hostPort string, id string) DynamoServer {
	preferenceList := make([]DynamoNode, 0)
	dataStore := make(map[string][]ObjectEntry)
	gossipList := make(map[int]PutArgs)
	selfNodeInfo := DynamoNode{
		Address: hostAddr,
		Port:    hostPort,
	}
	return DynamoServer{
		wValue:         w,
		rValue:         r,
		preferenceList: preferenceList,
		selfNode:       selfNodeInfo,
		nodeID:         id,
		dataStore:      dataStore,
		crashed:        false,
		gossipList:     gossipList,
	}
}

func ServeDynamoServer(dynamoServer DynamoServer) error {
	rpcServer := rpc.NewServer()
	e := rpcServer.RegisterName("MyDynamo", &dynamoServer)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Name Registration")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Registered the RPC Interfaces")

	l, e := net.Listen("tcp", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Port Listening")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Listening to Target Port ", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	log.Println(DYNAMO_SERVER, "Serving Server Now")

	return http.Serve(l, rpcServer)
}
