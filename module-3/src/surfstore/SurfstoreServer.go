package surfstore

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Server struct {
	BlockStore BlockStoreInterface
	MetaStore  MetaStoreInterface
}

func (s *Server) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
	err := s.MetaStore.GetFileInfoMap(succ, serverFileInfoMap)
	return err
}

func (s *Server) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
	err := s.MetaStore.UpdateFile(fileMetaData, latestVersion)
	return err
}

func (s *Server) GetBlock(blockHash string, blockData *Block) error {
	err := s.BlockStore.GetBlock(blockHash, blockData)
	return err
}
func (s *Server) PutBlock(blockData Block, succ *bool) error {
	err := s.BlockStore.PutBlock(blockData, succ)
	return err
}
func (s *Server) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	err := s.BlockStore.HasBlocks(blockHashesIn, blockHashesOut)
	return err
}

// This line guarantees all method for surfstore are implemented
var _ Surfstore = new(Server)

func NewSurfstoreServer() Server {
	blockStore := BlockStore{BlockMap: map[string]Block{}}
	metaStore := MetaStore{FileMetaMap: map[string]FileMetaData{}}

	return Server{
		BlockStore: &blockStore,
		MetaStore:  &metaStore,
	}
}

func ServeSurfstoreServer(hostAddr string, surfstoreServer Server) error {
	rpc.Register(&surfstoreServer)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", hostAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	log.Printf("Serving rpc on port %s", hostAddr)

	return http.Serve(listener, nil)
}
