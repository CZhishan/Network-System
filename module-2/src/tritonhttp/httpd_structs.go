package tritonhttp


type HttpServer	struct {
	ServerPort	string
	DocRoot		string
	MIMEPath	string
	MIMEMap		map[string]string
}

type HttpResponseHeader struct {
	// Add any fields required for the response here
	Server				string
	Last_Modified		string
	Content_Type		string
	Content_Length		int64
	Connection			string
	Path   				string

}

type HttpRequestHeader struct {
	// Add any fields required for the request here
	Valid			bool
	URL				string
	Host			string
	Connection		string
}
