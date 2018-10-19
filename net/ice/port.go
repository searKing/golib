package ice

var portMap = map[string]string{
	"stun":  "3478",
	"turn":  "3478",
	"stuns": "5349",
	"turns": "5349",
}
var getDefaultPort = func(schema string) string {
	port, ok := portMap[schema]
	if ok {
		return port
	}
	return ""
}
