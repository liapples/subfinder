// 
// crtsh.go : A Golang based client for CRT.SH Parsing
// Written By : @ice3man (Nizamul Rana)
// 
// Distributed Under MIT License
// Copyrights (C) 2018 Ice3man
//

package crtsh

import (
	"io/ioutil"
	"strings"
	"encoding/json"
	"fmt"

	"subfinder/libsubfinder/helper"
)

// Structure of a single dictionary of output by crt.sh
// We only need name_value object hence this :-)
type crtsh_object struct {
	Name_value		string  `json:"name_value"`
}

// array of all results returned
var crtsh_data []crtsh_object

// all subdomains found
var subdomains []string 

// 
// Query : Queries awesome crt.sh service by comodo
// @param state : current application state, holds all information found
//
func Query(state *helper.State, ch chan helper.Result) {

	var result helper.Result
	result.Subdomains = subdomains

	// Make a http request to CRT.SH server and request output in JSON
	// format.
	// I Think 5 minutes would be more than enough for CRT.SH :-)
	resp, err := helper.GetHTTPResponse("https://crt.sh/?q=%25."+state.Domain+"&output=json", 3000)
	if err != nil {
		result.Error = err
		ch <- result
	}

	// Get the response body
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		ch <- result
	}

	if strings.Contains(string(resp_body), "The requested URL / was not found on this server.") {
		// crt.sh is not showing subdomains for some reason
		// move back
		result.Error = nil
		ch <- result
	}

	// Convert Response Body to string and then replace }{ to },{
	// This is done in order to enable parsing of JSON format employed by
	// crt.sh
	correct_format := strings.Replace(string(resp_body), "}{", "},{", -1)

	// Now convert it into a json array like this
	// [
	// 		{abc},
	//		{abc}
	// ]
	json_output := "["+correct_format+"]"

	// Decode the json format
	err = json.Unmarshal([]byte(json_output), &crtsh_data)
	if err != nil {
		result.Error = err
		ch <- result
	}

	// Append each subdomain found to subdomains array
	for _, subdomain := range crtsh_data {

		// Fix Wildcard subdomains containg asterisk before them
		if strings.Contains(subdomain.Name_value, "*.") {
			subdomain.Name_value = strings.Split(subdomain.Name_value, "*.")[1]
		}

		if state.Verbose == true {
			if state.Color == true {
				fmt.Printf("\n[%sCRT.SH%s] %s", helper.Red, helper.Reset, subdomain.Name_value)
			} else {
				fmt.Printf("\n[CRT.SH] %s", subdomain.Name_value)
			}
		}

		subdomains = append(subdomains, subdomain.Name_value)
	}

	result.Subdomains = subdomains
	result.Error = nil
	ch <-result
}