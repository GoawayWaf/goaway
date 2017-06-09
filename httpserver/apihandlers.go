package httpserver

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"justdevelop.it/goaway/repos/verdict"
	"justdevelop.it/goaway/utils"
	"net"
	"net/http"
	"strconv"
	"time"
)

func serveResp(w http.ResponseWriter, resp *jsonResponse) {

	setJsonheader(w)
	if len(resp.Errors) > 0 && resp.Errors[0].Status > 0 {

		interr := resp.Errors[0].Status
		fmt.Println(resp.Errors[0])
		w.WriteHeader(interr)
	}  else {
		w.WriteHeader(http.StatusOK)
	}
	err := json.NewEncoder(w).Encode(resp)
	utils.CheckAndPanic(err)
}

func IpRuleList(w http.ResponseWriter, r *http.Request) {
	resp := &jsonResponse{}
	rules, err := verdictRepo.FetchRules()

	if err != nil {

		resp.AddError(&respError{Status: http.StatusInternalServerError, Title: "Error Getting Data", Details: err.Error()})
	}

	for _, v := range rules {
		ip := utils.Ip2int(net.ParseIP(v.Ip))
		data := &respData{Id: ip, DataType: "ip", Attributes: v}
		resp.AddData(data)
	}
	serveResp(w, resp)
}

func IpRule(w http.ResponseWriter, r *http.Request) {
	resp := &jsonResponse{}

	vars := mux.Vars(r)
	intIp, err := strconv.ParseUint(vars["id"], 10, 0)

	rules, err := verdictRepo.FetchRules()
	if err != nil {
		respErr := &respError{Status: http.StatusInternalServerError, Title: "Error Fetching ip", Details: err.Error()}
		resp.AddError(respErr)
	} else {
		var rule verdict.IpSentence
		for _, v := range rules {
			if uint32(intIp) == v.IpInt {
				rule = v
			}
		}
		if rule.IpInt > 0 {
			data := &respData{Id: intIp, DataType: "ip", Attributes: rule}
			resp.AddData(data)

		} else {
			respErr := &respError{Status: http.StatusNotFound, Title: "Ip not found"}
			resp.AddError(respErr)
		}

	}
	serveResp(w, resp)
}

/*
Test curl:
curl -H "Content-Type: application/json" -d '{"name":"New Todo"}' http://localhost:8085/iprule
*/
func IpRuleCreate(w http.ResponseWriter, r *http.Request) {
	resp := &jsonResponse{}
	var sentence verdict.IpSentence
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &sentence); err != nil {
		respErr := &respError{Status: http.StatusInternalServerError, Title: "Creation error", Details: err.Error()}
		resp.AddError(respErr)
	} else {

		ip := net.ParseIP(sentence.Ip)
		if ip == nil {
			respErr := &respError{Status: http.StatusInternalServerError, Title: "Ip Not valid", Details: sentence.Ip}
			resp.AddError(respErr)
		} else {
			sentence.IpInt = utils.Ip2int(net.ParseIP(sentence.Ip))
			sentence.BannedBy = "api"
			sentence.DateTime = time.Now().Format(utils.MYSQLDATETIME)
			id, err := verdictRepo.Store(&sentence)
			fmt.Println(id)
			w.WriteHeader(http.StatusCreated)
			if err != nil {
				respErr := &respError{Status: http.StatusInternalServerError, Title: "Creation error", Details: err.Error()}
				resp.AddError(respErr)
			} else {
				data := &respData{Id: sentence.IpInt, DataType: "ip", Attributes: sentence}
				resp.AddData(data)
			}
		}
	}
	serveResp(w, resp)
}

func setJsonheader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}
