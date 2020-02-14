package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

func main() {

	log.Println("Starting up...")

	url, exists := os.LookupEnv("URL_LABEL")

	if exists {

		log.Println("Checking for: ", url)
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/swarm-nodes", numberOfSwarmNodes)
		router.HandleFunc("/swarm-services", listServices)
		log.Fatal(http.ListenAndServe(":7001", router))

	}

}

func numberOfSwarmNodes(w http.ResponseWriter, r *http.Request) {

	log.Println("numberOfSwarmNodes")

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))

	if err != nil {
		panic(err)
	}

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		panic(err)
	}

	//return len(swarmNodes)

	htmlOutput := "<html>"
	htmlOutput += "" + strconv.Itoa(len(swarmNodes)) + "<br/>"
	htmlOutput += "</html>"
	fmt.Fprint(w, htmlOutput)

	log.Println("Called numberOfSwarmNodes")

}

func totalSwarmNodes(cli *client.Client) int {

	log.Println("totalSwarmNodes")

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		panic(err)
	}

	log.Println("Called totalSwarmNodes")

	return len(swarmNodes)

}

func listServices(w http.ResponseWriter, r *http.Request) {

	log.Println("listServices")

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))

	if err != nil {
		panic(err)
	}

	//List all Swarm services
	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	nodes := totalSwarmNodes(cli)

	htmlOutput := "<html>"
	for _, service := range services {
		sgURL := ""
		for k, v := range service.Spec.Labels {
			if k == os.Getenv("URL_LABEL") {
				sgURL = v
			}
		}
		modeStr := ""
		replicas := ""
		if service.Spec.Mode.Global != nil {
			modeStr = "Global"
			replicas = strconv.Itoa(nodes)
		}
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			modeStr = "Replicated"
			replicas = strconv.FormatUint(*service.Spec.Mode.Replicated.Replicas, 10)
		}

		htmlOutput += fmt.Sprintf("%s | %s | %s | %s | %s | %v\n", service.ID, service.Spec.Name, sgURL, modeStr, replicas, service.Endpoint.Ports)
		htmlOutput += "<br/>"

	}
	htmlOutput += "</html>"
	fmt.Fprint(w, htmlOutput)
}
