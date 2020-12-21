package surfstore
import(
	"crypto/sha256"
	"errors"
	"encoding/hex"
	//"fmt"
)
type BlockStore struct {
	BlockMap map[string]Block
}

func (bs *BlockStore) GetBlock(blockHash string, blockData *Block) error {
	if value, ok := bs.BlockMap[blockHash]; ok {
	   	*blockData = value
			return nil
	}
	return errors.New("The Block doesn't exist!")
}

func (bs *BlockStore) PutBlock(block Block, succ *bool) error {
	var hashCode string
	h := sha256.Sum256(block.BlockData)
	hashCode = hex.EncodeToString(h[:])
	if _, ok := bs.BlockMap[hashCode]; !ok {
		bs.BlockMap[hashCode] = block
		//fmt.Println("Upload a block!")
	}
	*succ = true
	return nil
}

func (bs *BlockStore) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	for _, hash := range blockHashesIn {
		if _, ok := bs.BlockMap[hash]; ok {
			*blockHashesOut = append(*blockHashesOut, hash)
		}
	}
	return nil
}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)
