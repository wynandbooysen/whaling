package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

type Container struct {
	ServiceID     string `json:"serviceId"`
	Name          string `json:"name"`
	URL           string `json:"URL"`
	RepMode       string `json:"repMode"`
	Replicas      string `json:"replicas"`
	PublishedPort string `json:"pubPort"`
}

var containers = []Container{}

func main() {

	log.Println("Starting up...")

	url, exists := os.LookupEnv("URL_LABEL")

	if exists {

		log.Println("Checking for:", url)
		router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/swarm-nodes", numberOfSwarmNodes)
		router.HandleFunc("/swarm-services", listServices)
		router.HandleFunc("/swarm-services-json", jsonServices)
		log.Fatal(http.ListenAndServe(":7002", router))

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
		URL := ""
		for k, v := range service.Spec.Labels {
			if k == os.Getenv("URL_LABEL") {
				URL = v
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

		//Get published port number
		portNumber := ""
		for _, port := range service.Endpoint.Ports {

			if port.Protocol == "tcp" {

				portNumber = portNumber + fmt.Sprint(port.PublishedPort) + ","

			}
		}
		portNumber = strings.TrimSuffix(portNumber, ",")

		htmlOutput += fmt.Sprintf("%s | %s | %s | %s | %s | %v\n", service.ID, service.Spec.Name, URL, modeStr, replicas, portNumber)
		htmlOutput += "<br/>"

	}
	htmlOutput += "</html>"
	fmt.Fprint(w, htmlOutput)
}

func jsonServices(w http.ResponseWriter, r *http.Request) {

	log.Println("JSON")

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
	//clear containers slice to prevent duplication
	containers = nil

	for _, service := range services {
		URL := ""
		for k, v := range service.Spec.Labels {
			if k == os.Getenv("URL_LABEL") {
				URL = v
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

		//Get published port number
		portNumber := ""
		for _, port := range service.Endpoint.Ports {

			if port.Protocol == "tcp" {

				portNumber = portNumber + fmt.Sprint(port.PublishedPort) + ","

			}
		}
		portNumber = strings.TrimSuffix(portNumber, ",")

		newContainer := Container{
			ServiceID:     service.ID,
			Name:          service.Spec.Name,
			URL:           URL,
			RepMode:       modeStr,
			Replicas:      replicas,
			PublishedPort: portNumber,
		}
		containers = append(containers, newContainer)

	}
	json.NewEncoder(w).Encode(containers)
}
