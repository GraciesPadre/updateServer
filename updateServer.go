package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

const port = 8080

var latestUpdateRev = "gracie_0"
var networkInterfaceName = "en0"

// On linux with a wireless network, start with -networkInterfaceName wlp0s20f3
// curl -k https://localhost:8080/update/2/latest | json_pp
// curl -k https://10.0.0.28:8080/gracie/r_gracie_02.bin

func main() {
	useHttp := false

	for i, arg := range os.Args[1:] {
		if arg == "-useHttp" {
			useHttp = true
		}

		if arg == "-reportOldUpdateVersion" {
			latestUpdateRev = "gracie_01"
		}

		if arg == "-networkInterfaceName" {
			if len(os.Args) > i+2 {
				// We're looking at args starting at index 1, so, to get the next arg, we need to increment
				// the current arg index and add 1 to account for not starting at 0.
				networkInterfaceName = os.Args[i+2]
			}
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/update/{model}/latest", updateMetaDataHandler)
	router.HandleFunc("/gracie/{fileName}", FileDownloadHandler)

	fmt.Println("Starting update server on network interface:", networkInterfaceName, "port:", port)

	if useHttp {
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
	} else {
		log.Fatalln(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "localhost.crt", "localhost.key", router))
	}
}

func updateMetaDataHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	vars := mux.Vars(httpRequest)
	modelText, ok := vars["model"]

	if !ok {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte("No model id in request\n"))
		if err != nil {
			fmt.Println("Could not write no model id response")
		}

		return
	}

	model, err := strconv.ParseInt(modelText, 10, 32)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte(fmt.Sprintf("Could not convert %s to a model id\n", modelText)))
		if err != nil {
			fmt.Println("Could not write model id conversion response")
		}

		return
	}

	fmt.Printf("Model id: %d\n", model)

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting interfaces")
		return
	}

	var ipAddress string

	// Find the network interface name that houses the primary connection, to be able
	// to get its ipv4 ip address.  That ip address is used in the response payload from
	// updateMetaDataHandler to let clients know the update file url.  That address
	// should also be the ip address used to generate https certificates in the CN field,
	// e.g. openssl req -new -subj "/C=US/ST=Colorado/CN=10.0.0.6" -newkey rsa:2048 -nodes -keyout localhost.key -out localhost.csr
	// Not all clients need the ip address specified this way, but some browsers
	// won't accept a certificate, unless it has a CN field that matches the url the
	// browser connected to.
	for _, iFace := range interfaces {
		if iFace.Name == networkInterfaceName {
			ipAddresses, err := iFace.Addrs()

			if err != nil {
				continue
			}

			for _, addr := range ipAddresses {
				switch v := addr.(type) {
				case *net.IPNet:
					ipV4Addr := v.IP.To4()
					if ipV4Addr != nil {
						ipAddress = ipV4Addr.String()
					}
				}
			}
		}
	}

	if len(ipAddress) == 0 {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte("Could not get IP address\n"))
		if err != nil {
			fmt.Println("Could not write IP address response")
		}

		return
	}

	updateMetaData := NewUpdateMetaData(
		fmt.Sprintf("https://%s:%d/gracie/r_%s%d.bin", ipAddress, port, latestUpdateRev, model),
		fmt.Sprintf("%s%d", latestUpdateRev, model),
	)
	responsePayload, err := json.Marshal(updateMetaData)
	responsePayload = append(responsePayload, []byte("\n")...)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte("Could not marshal response data\n"))
		if err != nil {
			fmt.Println("Could not write marshal failure")
		}

		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	_, err = responseWriter.Write(responsePayload)
	if err != nil {
		fmt.Printf("Failed to write response: %s\n", err)
	}
}

func FileDownloadHandler(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	vars := mux.Vars(httpRequest)
	fileName, ok := vars["fileName"]

	if !ok {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte("No file name in request\n"))
		if err != nil {
			fmt.Println("Could not write no file name response")
		}

		return
	}

	fileData, err := os.ReadFile(fileName)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		_, err := responseWriter.Write([]byte("Could not read file\n"))
		if err != nil {
			fmt.Println("Could not write read file response")
		}

		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Header().Set("Content-Type", "application/octet-stream")
	_, err = responseWriter.Write(fileData)
	if err != nil {
		fmt.Println("Writing result failed")
	}
}
