package api

import (
	"net/http"
	"os"
	"strings"
)


var RootPath string


func init() {
	rootPath, rootPathAvailable := os.LookupEnv("ROOT_PATH")
	if rootPathAvailable && rootPath != "" {
		RootPath = rootPath
	}
}

// SetupRoutes registers the /v1/ endpoints with the http package.
func SetupRoutes() {
	http.HandleFunc(RootPath+"/v1/stats", statsEndpoint)
	//The websocket is shared between the public API and the official client
	http.HandleFunc(RootPath+"/v1/ws", wsEndpoint)

	//These exist only for the public API. We version them in order to ensure
	//backwards compatibility as far as possible.
	http.HandleFunc(RootPath+"/v1/lobby", lobbyEndpoint)
	http.HandleFunc(RootPath+"/v1/lobby/player", enterLobbyEndpoint)
}

func remoteAddressToSimpleIP(input string) string {
	address := input
	lastIndexOfDoubleColon := strings.LastIndex(address, ":")
	if lastIndexOfDoubleColon != -1 {
		address = address[:lastIndexOfDoubleColon]
	}

	return strings.TrimSuffix(strings.TrimPrefix(address, "["), "]")
}


func GetIPAddressFromRequest(r *http.Request) string {

	forwardedAddress := r.Header.Get("X-Forwarded-For")
	if forwardedAddress != "" {
		clientAddress := strings.TrimSpace(strings.Split(forwardedAddress, ",")[0])
		return remoteAddressToSimpleIP(clientAddress)
	}

	standardForwardedHeader := r.Header.Get("Forwarded")
	if standardForwardedHeader != "" {
		targetPrefix := "for="
		//Since forwarded can contain more than one field, we search for one specific field.
		for _, part := range strings.Split(standardForwardedHeader, ";") {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(trimmed, targetPrefix) {
				//FIXME Maybe checking for a valid IP-Address would make sense here, not sure tho.
				address := remoteAddressToSimpleIP(strings.TrimPrefix(trimmed, targetPrefix))
				//Since the documentation doesn't mention which quotes are used, I just remove all ;)
				return strings.NewReplacer("`", "", "'", "", "\"", "", "[", "", "]", "").Replace(address)
			}
		}
	}

	return remoteAddressToSimpleIP(r.RemoteAddr)
}
