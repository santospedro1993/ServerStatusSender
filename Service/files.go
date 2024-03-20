package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)
main

import (
"encoding/json"
"fmt"
"io/ioutil"
"os"
)

type Config struct {
	Values map[string]interface{}
}

func CreateFile() {
	c := Config{Values: map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}}

	file, _ := json.MarshalIndent(c, "", " ")

	_ = ioutil.WriteFile("example.json", file, 0644)
}

func readFile() {
	file, _ := ioutil.ReadFile("example.json")

	data := Config{}

	_ = json.Unmarshal([]byte(file), &data)

	fmt.Println(data.Values["key1"])
	fmt.Println(data.Values["key2"])
	fmt.Println(data.Values["key3"])
}

func main() {
	CreateFile()
	readFile()
}

func main() {

}
