package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// create a handler struct
type HttpHandler struct{}

//create the structure according to the request recieve
type test_struct struct {
	Ev     string
	Et     string
	Id     string
	Uid    string
	Mid    string
	T      string
	P      string
	L      string
	Sc     string
	Atrk1  string
	Atrv1  string
	Atrt1  string
	Atrk2  string
	Atrv2  string
	Atrt2  string
	Uatrk1 string
	Uatrv1 string
	Uatrt1 string
	Uatrk2 string
	Uatrv2 string
	Uatrt2 string
	Uatrk3 string
	Uatrv3 string
	Uatrt3 string
}

//define the structure to store json type
type Attributes struct {
	Form_varient map[string]string
	Ref          map[string]string
}

type Traits struct {
	Name  map[string]string
	Email map[string]string
	Age   map[string]string
}

type Datas struct {
	Event, Event_type, App_id, User_id, Message_id, Page_title, Page_url, Browser_language, Screen_size string
	Attributes                                                                                          Attributes
	Traits                                                                                              Traits
}

// implement `ServeHTTP` method on `HttpHandler` struct
func (h HttpHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	var t test_struct
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	v := reflect.ValueOf(t)
	typeOfS := v.Type()

	//create the channel send this request to worker
	c := make(chan map[string]string)
	for i := 0; i < v.NumField(); i++ {
		key := typeOfS.Field(i).Name
		value := v.Field(i).Interface()
		val, _ := value.(string)
		go worker(c, key, val)
	}
	//parse the data send it to worker
	go worker2(c, t)

	// create response binary data
	data := []byte("Success!")

	// write `data` to response
	res.Write(data)
}

//create the worker to make this into map and return
func worker(c chan map[string]string, key string, value string) {

	var retmap = make(map[string]string)
	//	retmap := make(map[string][]string)
	retmap[key] = value
	c <- retmap
}

//declare worker2 to parse the data from map and make json format
func worker2(c chan map[string]string, t test_struct) {

	//parse the data from the channel and store in map
	value := reflect.ValueOf(t)
	var output Datas
	result := make(map[string]string)

	for i := 0; i < value.NumField(); i++ {
		m := <-c
		for k, v := range m {
			result[k] = v
		}
	}
	//form the json format
	output = Datas{
		Event:            result["Ev"],
		Event_type:       result["Et"],
		App_id:           result["Id"],
		User_id:          result["Uid"],
		Message_id:       result["Mid"],
		Page_title:       result["T"],
		Page_url:         result["P"],
		Browser_language: result["L"],
		Screen_size:      result["Sc"],
		Attributes: Attributes{
			Form_varient: map[string]string{"value": result["Atrv1"], "type": result["Atrt1"]},
			Ref:          map[string]string{"value": result["Atrv2"], "type": result["Atrt2"]},
		},
		Traits: Traits{
			Name:  map[string]string{"value": result["Uatrv1"], "type": result["Uatrt1"]},
			Age:   map[string]string{"value": result["Uatrv2"], "type": result["Uatrt2"]},
			Email: map[string]string{"value": result["Uatrv3"], "type": result["Uatrt3"]},
		},
	}

	//encode `output` as JSON
	dataJSON, _ := json.MarshalIndent(output, "", "  ")
	//post the json to the url
	//https://webhook.site/#!/1ce4e5f8-87da-4920-aa9a-c4ebc03b1adb
	resp, _ := http.Post("https://webhook.site/1ce4e5f8-87da-4920-aa9a-c4ebc03b1adb/", "/", bytes.NewBuffer(dataJSON))

	fmt.Println(string(dataJSON))
	fmt.Println(resp)

}

func main() {

	// create a new handler
	handler := HttpHandler{}

	// listen and serve
	http.ListenAndServe(":9000", handler)

}
