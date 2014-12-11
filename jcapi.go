package jcapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	responseSize = 256 * 1024
	stdUrlBase   = "https://console.jumpcloud.com/api"
)

type JCOp uint8

const (
	read   = 1
	insert = 2
	update = 3
	delete = 4
	list   = 5
)

type JCAPI struct {
	ApiKey  string
	UrlBase string
}

type JCError interface {
	Error() string
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func newJCAPI(apiKey string, urlBase string) JCAPI {
	return JCAPI{
		ApiKey:  apiKey,
		UrlBase: urlBase,
	}
}

func buildJSONStringArray(field string, s []string) string {
	returnVal := "["

	if s != nil {
		afterFirst := false

		for _, val := range s {
			if afterFirst {
				returnVal += ","
			}

			returnVal += "\"" + val + "\""

			afterFirst = true
		}
	}
	returnVal += "]"

	return "\"" + field + "\":" + returnVal
}

func buildJSONKeyValuePair(key, value string) string {
	return "\"" + key + "\":\"" + value + "\""
}

func buildJSONKeyValueBoolPair(key string, value bool) string {
	if value == true {
		return "\"" + key + "\":\"true\""
	} else {
		return "\"" + key + "\":\"false\""
	}

}

func getTimeString() string {
	t := time.Now()

	return t.Format(time.RFC3339)
}

func (jc JCAPI) emailFilter(email string) []byte {

	//
	// Ideally, this would be generalized to take a map[string]string
	// but, that doesn't elicit the correct JSON output for the JumpCloud
	// filters in json.Marshal()
	//
	return []byte(fmt.Sprintf("{\"filter\": [{\"email\" : \"%s\"}]}", email))
}

func (jc JCAPI) setHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", jc.ApiKey)
}

func (jc JCAPI) post(url string, data []byte) (interface{}, JCError) {
	return jc.do(mapJCOpToHTTP(insert), url, data)
}

func (jc JCAPI) put(url string, data []byte) (interface{}, JCError) {
	return jc.do(mapJCOpToHTTP(update), url, data)
}

func (jc JCAPI) delete(url string) (interface{}, JCError) {
	return jc.do(mapJCOpToHTTP(delete), url, nil)
}

func (jc JCAPI) get(url string) (interface{}, JCError) {
	return jc.do(mapJCOpToHTTP(read), url, nil)
}

func (jc JCAPI) list(url string) (interface{}, JCError) {
	return jc.do(mapJCOpToHTTP(list), url, nil)
}

func (jc JCAPI) do(op, url string, data []byte) (interface{}, JCError) {
	var returnVal interface{}

	fullUrl := jc.UrlBase + url

	client := &http.Client{}

	dbg(3, "JCAPI.do(): op='%s' - url='%s' - data='%s'\n", op, fullUrl, data)

	req, err := http.NewRequest(op, fullUrl, bytes.NewReader(data))
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not build search request: '%s'", err)
	}

	jc.setHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: client.Do() failed, err='%s'", err)
	}

	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return returnVal, fmt.Errorf("JumpCloud HTTP response status='%s'", resp.Status)
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not read the response body, err='%s'", err)
	}

	dbg(3, "200 OK Response Buffer=%U\n", string(buffer))

	err = json.Unmarshal(buffer, &returnVal)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not Unmarshal JSON response, err='%s'", err)
	}

	return returnVal, err
}

// Add all the tags of which the user is a part to the JCUser object
func (user *JCUser) addTags(tags []JCTag) {
	for _, tag := range tags {
		for _, systemUser := range tag.SystemUsers {
			if systemUser == user.Id {
				user.tags = append(user.tags, tag)
			}
		}
	}
}

func mapJCOpToHTTP(op JCOp) string {
	var returnVal string

	switch op {
	case read:
		returnVal = "GET"
	case insert:
		returnVal = "POST"
	case update:
		returnVal = "PUT"
	case delete:
		returnVal = "DELETE"
	case list:
		returnVal = "LIST"
	}

	return returnVal
}

//
// Interface Conversion Helper Functions
//
func (jc JCAPI) extractStringArray(input []interface{}) []string {
	var returnVal []string

	for _, str := range input {
		returnVal = append(returnVal, str.(string))
	}

	return returnVal
}

func getStringOrNil(input interface{}) string {
	returnVal := ""

	switch input.(type) {
	case string:
		returnVal = input.(string)
	}

	return returnVal
}

func getUint16OrNil(input interface{}) uint16 {
	var returnVal uint16

	switch input.(type) {
	case uint16:
		returnVal = input.(uint16)
	}

	return returnVal
}
