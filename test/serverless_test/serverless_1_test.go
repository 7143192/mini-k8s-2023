package serverless

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"
)

func ServerlessTrigger() {
	url := ""
	flag := 1
	if flag == 1 {
		url = "http://localhost:8081/objectAPI/triggerWorkFlow/workflow-example"
	} else {
		url = "http://localhost:8081/objectAPI/trigger/print_high"
	}
	params := make(map[string]string)
	params["x"] = "19980614"
	body, err := json.Marshal(params)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	result := make(map[string][]byte)
	err = json.NewDecoder(res.Body).Decode(&result)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	if res.StatusCode == http.StatusOK {
		data := make(map[string]string)

		err = json.Unmarshal(result["INFO"], &data)

		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}

		log.Printf("[INFO] %v\n", data)
	} else {
		log.Printf("[ERROR] %v\n", string(result["ERROR"]))
	}
}

func TestServerless(t *testing.T) {
	for i := 0; i < 10; i++ {
		go ServerlessTrigger()
	}
	for {
		time.Sleep(5 * time.Second)
	}
}
