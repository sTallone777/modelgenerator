package conf

// model files path
const ModelPath = "./models/"

// "" or "*"
const Pointer = ""

type DbConf struct {
	Host   string
	Port   string
	User   string
	Pwd    string
	DbName string
}

var DbConfig = DbConf{
	Host:   "127.0.0.1",
	Port:   "1433",
	User:   "",
	Pwd:    "",
	DbName: "",
}
