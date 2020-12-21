package surfstore
import(
	"errors"
	"log"
)

type MetaStore struct {
	FileMetaMap map[string]FileMetaData
}

func (m *MetaStore) GetFileInfoMap(_ignore *bool, serverFileInfoMap *map[string]FileMetaData) error {
	*serverFileInfoMap = m.FileMetaMap
	return nil
}

func (m *MetaStore) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) (err error) {
	filename := fileMetaData.Filename
	if file, ok := m.FileMetaMap[filename]; ok {
		if fileMetaData.Version - m.FileMetaMap[filename].Version == 1{
			file.Version = fileMetaData.Version
			file.BlockHashList = fileMetaData.BlockHashList
			m.FileMetaMap[filename] = file
			*latestVersion = fileMetaData.Version
			return nil
		} else if m.FileMetaMap[filename].Version < fileMetaData.Version{
			err = errors.New("Your version is too new!")
		} else if m.FileMetaMap[filename].Version == fileMetaData.Version{
			err = errors.New("Same Version!")
		} else{
			err = errors.New("Your file is too old!")
		}
		*latestVersion = m.FileMetaMap[filename].Version
		log.Println(err)
		return err
		
	} else {
		var newfile FileMetaData
		newfile.Filename = filename
		newfile.Version = fileMetaData.Version
		newfile.BlockHashList = fileMetaData.BlockHashList
		m.FileMetaMap[filename] = newfile
		*latestVersion = fileMetaData.Version
		return nil
	}

}

var _ MetaStoreInterface = new(MetaStore)
