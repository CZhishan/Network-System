package tritonhttp

import(
	"fmt"
  "os"
  "bufio"
  "strings"
)
/** 
	Load and parse the mime.types file 
**/
func ParseMIME(MIMEPath string) (MIMEMap map[string]string, err error) {
	mime_map, err := os.Open(MIMEPath)
	if err != nil {
		fmt.Println(mime_map)
	}
	scanner := bufio.NewScanner(mime_map)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan(){
		txtlines = append(txtlines,scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	defer func(){
		if err = mime_map.Close();err != nil {
			panic(err)
		}
	}()
	mime := make(map[string]string)
	for _, line := range txtlines{
		new_line := strings.Fields(line)
		mime[new_line[0]] = new_line[1]
		//fmt.Println(new_line)
	}
	return mime,nil
}

