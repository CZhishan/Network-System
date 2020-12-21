package surfstore

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"crypto/sha256"
	"encoding/hex"
	"bufio"
	"reflect"
	"strings"
	"strconv"
)

/*
Implement the logic for a client syncing with the server here.
*/
func ClientSync(client RPCClient) {
	fileMap := make(map[string][]string)

	files, _ := ioutil.ReadDir(client.BaseDir)
    for _, fi := range files {
        if fi.Name() == "index.txt" {
			continue
        }
		file, _ := os.Open(client.BaseDir + "/" + fi.Name())
		buf := make([]byte, client.BlockSize)
		//var hashlist []string
		for {
			count, err := file.Read(buf)
			if err == io.EOF {
				break;
			}
			currBytes := buf[:count]
			h := sha256.Sum256(currBytes)
			hashCode := hex.EncodeToString(h[:])
			fileMap[fi.Name()] = append(fileMap[fi.Name()], hashCode)
		}
		file.Close()
	}
	//read the local Index
	localIndex,err := readIndex(client)
	if err!=nil{
		fmt.Println(err)
	}

	p := new(bool)
	var serverMap map[string]FileMetaData
	client.GetFileInfoMap(p, &serverMap)
	//PrintMetaMap(serverMap)
	//for every file in the local directory
	for filename, hashlist := range fileMap{
		if localInfo, ok := localIndex[filename]; ok {
			//if there are local modifications to the file
			if !reflect.DeepEqual(localInfo.BlockHashList, hashlist) {
				if serverMap[filename].Version == localInfo.Version {
					UpdateRemoteFile(client, filename)
					// Update filemetadata and handle conflicts
					var latestV int
					newInfo := FileMetaData{Filename: filename, Version: localInfo.Version + 1, BlockHashList: hashlist}
					err := client.UpdateFile(&newInfo, &latestV)
					if err == nil {
						localIndex[filename] = newInfo
					} else {
						p := new(bool)
						var newServerMap map[string]FileMetaData
						client.GetFileInfoMap(p, &newServerMap)
						if newServerMap[filename].Version > localInfo.Version {
							UpdateLocalFile(client, filename, newServerMap[filename])
							localIndex[filename] = newServerMap[filename]
						}
						
					}
				} 
			} 
			// remote version larger than local version
			if serverMap[filename].Version > localInfo.Version {
				UpdateLocalFile(client, filename, serverMap[filename])
				localIndex[filename] = serverMap[filename]
			}
		// new local files(not in the local index)
		} else {
			UpdateRemoteFile(client, filename)
			// Update filemetadata and handle conflicts
			var latestV int
			newInfo := FileMetaData{Filename: filename, Version: 1, BlockHashList: hashlist}
			err := client.UpdateFile(&newInfo, &latestV)
			if err == nil {
				localIndex[filename] = newInfo
			} else {
				p := new(bool)
				var newServerMap map[string]FileMetaData
				client.GetFileInfoMap(p, &newServerMap)
				if newServerMap[filename].Version >= 1 {
					UpdateLocalFile(client, filename, newServerMap[filename])
					localIndex[filename] = newServerMap[filename]
				}
			}
		}
	}

	// handle delete files
	for filename, localInfo := range localIndex {
		if localInfo.BlockHashList[0] == "0" {
			continue
		}
		if _, ok := fileMap[filename]; !ok {
			var latestV int
			newInfo := FileMetaData{Filename: filename, Version: localInfo.Version + 1, BlockHashList: []string{"0"}}
			err := client.UpdateFile(&newInfo, &latestV)
			if err == nil {
				localIndex[filename] = newInfo
			} else {
				p := new(bool)
				var newServerMap map[string]FileMetaData
				client.GetFileInfoMap(p, &newServerMap)
				if newServerMap[filename].Version > localInfo.Version {
					UpdateLocalFile(client, filename, newServerMap[filename])
					localIndex[filename] = newServerMap[filename]
				}
			}
		}
	}

	// check new remote files
	for filename, remoteInfo := range serverMap {
		if remoteInfo.BlockHashList[0] == "0" {
			continue
		}
		//if the file in serverMap is not in the localIndex
		if _, ok := localIndex[filename]; !ok {
			UpdateLocalFile(client, filename, remoteInfo)
			localIndex[filename] = remoteInfo
		} else if _, ok = fileMap[filename]; !ok && remoteInfo.Version > localIndex[filename].Version {
			//recreate files
			UpdateLocalFile(client, filename, remoteInfo)
			localIndex[filename] = remoteInfo
		}
	}

	// write new index file
	//PrintMetaMap(localIndex)
	writeIndex(client, localIndex)
}



func UpdateLocalFile(client RPCClient, filename string, remoteInfo FileMetaData) error {
	f, err := os.OpenFile(client.BaseDir + "/" + filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, hash := range remoteInfo.BlockHashList {
		if hash == "0" {
			err = os.Remove(client.BaseDir + "/" + filename)
			if err != nil {
				fmt.Println("Delete error:", err)
			}
			break
		}
		var block Block
		client.GetBlock(hash, &block)
		f.Write(block.BlockData)
	}
	
	return nil
}

func UpdateRemoteFile(client RPCClient, filename string) {
	var succ bool
	var block Block
	block.BlockSize = client.BlockSize
	file, _ := os.Open(client.BaseDir + "/" + filename)
	buf := make([]byte, client.BlockSize)
	for {
		count, err := file.Read(buf)
		if err == io.EOF {
			break;
		}
		block.BlockData = buf[:count]
		client.PutBlock(block, &succ)
	}
	file.Close()
}


/*
Helper function to print the contents of the metadata map.
*/
func PrintMetaMap(metaMap map[string]FileMetaData) {

	fmt.Println("--------BEGIN PRINT MAP--------")

	for _, filemeta := range metaMap {
		fmt.Println("\t", filemeta.Filename, filemeta.Version, filemeta.BlockHashList)
	}

	fmt.Println("---------END PRINT MAP--------")

}

func readIndex(client RPCClient) (map[string]FileMetaData, error){
	localIndex := make(map[string]FileMetaData)
	if _, err := os.Stat(client.BaseDir+"/index.txt"); err != nil {
			if os.IsExist(err) {
				return localIndex,err
			}
			f ,err := os.Create(client.BaseDir+"/index.txt")
			if err != nil{
				return localIndex, err
			}
			f.Close()
	}
	f, _ := os.Open(client.BaseDir+"/index.txt")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan(){
		line := scanner.Text()
		s := strings.Split(line, ",")
		ver, _ := strconv.Atoi(s[1])
		md := FileMetaData{Filename: s[0], Version: ver, BlockHashList: strings.Split(s[2]," ")}
		localIndex[s[0]] = md
    }
	return localIndex,nil
}

func writeIndex(client RPCClient, localIndex map[string]FileMetaData) error {
	f, err := os.OpenFile(client.BaseDir + "/index.txt", os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
			fmt.Println(err)
			return err
	}   
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, info := range localIndex { 
			lineStr := fmt.Sprintf("%s,%s,%s", info.Filename, strconv.Itoa(info.Version), strings.Join(info.BlockHashList, " "))
			fmt.Fprintln(w, lineStr)
			//fmt.Println(lineStr)
	}   
	return w.Flush()
}
