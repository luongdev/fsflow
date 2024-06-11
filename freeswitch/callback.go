package freeswitch

//import (
//	"encoding/json"
//	"github.com/luongdev/fsflow/session/input"
//	"log"
//)
//
//func (e *Event) HasCallback() (*input.WorkflowCallback, bool) {
//	cbURL := e.GetHeader("variable_callback_url")
//	if cbURL == "" {
//		return nil, false
//	}
//
//	cbMethod := e.GetHeader("variable_callback_method")
//	if cbMethod == "" {
//		cbMethod = "POST"
//	}
//
//	cbHeaders := make(map[string]string)
//	cbHeadersStr := e.GetHeader("variable_callback_headers")
//	if cbHeadersStr == "" {
//		cbHeaders = map[string]string{"Content-Type": "application/json"}
//	} else {
//		if err := json.Unmarshal([]byte(cbHeadersStr), &cbHeaders); err != nil {
//			log.Printf("failed to unmarshal callback headers %v", err)
//			cbHeaders = map[string]string{"Content-Type": "application/json"}
//		}
//	}
//
//	cbBody := make(map[string]interface{})
//	cbBodyStr := e.GetHeader("variable_callback_body")
//	if cbBodyStr != "" {
//		if err := json.Unmarshal([]byte(cbBodyStr), &cbBody); err != nil {
//			log.Printf("failed to unmarshal callback body %v", err)
//		}
//	}
//
//	return &input.WorkflowCallback{URL: cbURL, Method: cbMethod, Body: cbBody, Headers: cbHeaders}, true
//}
