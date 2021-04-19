package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// create a handler struct
type HttpHandler struct{}

var POSTURL = "https://webhook.site/e3ca3185-0d17-4bae-95a0-c8a5fe4d38fe"
var PORT = ":9000"

// implement `ServeHTTP` method on `HttpHandler` struct
func handler(rw http.ResponseWriter, req *http.Request) {

	//Read the request body from http
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	//declare the interface map and unmarshal the request body
	var datas map[string]interface{}
	err = json.Unmarshal([]byte(body), &datas)
	if err != nil {
		panic(err)
	}

	//Create a channel and send the request in each go routine
	c := make(chan map[string]string)
	for key, value := range datas {
		val, _ := value.(string)
		go worker(c, key, val)
	}

	//pass the channel and unmarshal to worker 2 to form the json
	fmt.Println("Worker 2 start")
	go worker2(c, datas)

}

//create the worker to make this into map and return
func worker(c chan map[string]string, key string, value string) {

	var retmap = make(map[string]string)
	//	retmap := make(map[string][]string)
	retmap[key] = value
	c <- retmap
}

//Form the struct for inner json attribute
type JsonInnerFormat struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

//Declare the orderedmap to store the new map according to given json variable
type OrderedMap map[string]interface{}

//This function to form ordered json method
func (jsonmap OrderedMap) ToJson(sequence ...string) {

	buffer := &bytes.Buffer{}
	buffer.Write([]byte{'{', '\n'})
	//iterate the given sequence of map
	for _, key := range sequence {
		//check if the key is attribute or traits and form the inner json marshal
		if string(key) == "attribute" || string(key) == "traits" {
			formJson, _ := json.MarshalIndent(jsonmap[key], "\t", "  ")
			fmt.Fprintf(buffer, "\t\"%s\": %s", key, formJson)

			if string(key) == "attribute" {
				buffer.WriteByte(',')
			}
			//else make the json format from the key and map
		} else {
			fmt.Fprintf(buffer, "\t\"%s\": \"%v\"", key, jsonmap[key])
			buffer.WriteByte(',')
		}
		buffer.WriteByte('\n')
	}
	buffer.Write([]byte{'}', '\n'})

	//convert the *bytes to []bytes for http post
	readBuf, _ := ioutil.ReadAll(buffer)
	//post the json data to http://weebhook.site and print response
	print(string(readBuf))
	resp, _ := http.Post(POSTURL, "/", bytes.NewBuffer(readBuf))
	fmt.Println(resp)
}

//declare worker2 to parse the data from map and make json format
func worker2(c chan map[string]string, t map[string]interface{}) {

	result := make(map[string]string)
	atrkmap := make(map[string]JsonInnerFormat)
	uatrtmap := make(map[string]JsonInnerFormat)
	var uattr, attr = 0, 0

	//range the channel and form the new map
	for k := range t {
		print(k)
		m := <-c
		for key, value := range m {
			result[key] = value
			if strings.Contains(key, "atrk") && string(key[0]) == "a" {
				re := regexp.MustCompile("[0-9]+")
				get := re.FindAllString(key, -1)
				i, _ := strconv.Atoi(get[0])
				if i > attr {
					attr = i
				}
			} else if strings.Contains(key, "uatrk") {
				re := regexp.MustCompile("[0-9]+")
				get := re.FindAllString(key, -1)
				i, _ := strconv.Atoi(get[0])
				if i > uattr {
					uattr = i
				}
			}
		}
	}

	//for the attribute and traits from the map result
	for i := 1; i <= attr; i++ {
		get := strconv.Itoa(i)
		atrkmap[result["atrk"+get]] = JsonInnerFormat{Value: result["atrv"+get], Type: result["atrt"+get]}
	}

	for i := 1; i <= uattr; i++ {
		get := strconv.Itoa(i)
		uatrtmap[result["uatrk"+get]] = JsonInnerFormat{Value: result["uatrv"+get], Type: result["uatrt"+get]}

	}
	//form the ordered map and call the json and send the request to webhook
	jsonmap := OrderedMap{
		"event":            result["ev"],
		"event_type":       result["et"],
		"app_id":           result["id"],
		"user_id":          result["uid"],
		"message_id":       result["mid"],
		"page_title":       result["t"],
		"page_url":         result["p"],
		"browser_language": result["l"],
		"screen_size":      result["sc"],
		"attribute":        atrkmap,
		"traits":           uatrtmap}

	jsonmap.ToJson("event", "event_type", "app_id", "user_id", "message_id", "page_title", "page_url", "browser_language", "screen_size", "attribute", "traits")

}

func main() {

	//Create the handler
	http.HandleFunc("/json", handler)

	//listen the port
	log.Fatal(http.ListenAndServe(PORT, nil))
	fmt.Println("Done")

}
