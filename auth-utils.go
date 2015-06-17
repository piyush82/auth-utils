/*
 * Copyright (c) 2015. Zuercher Hochschule fuer Angewandte Wissenschaften
 *  All Rights Reserved.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License"); you may
 *     not use this file except in compliance with the License. You may obtain
 *     a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *     WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *     License for the specific language governing permissions and limitations
 *     under the License.
 */

/*
 *     Author: Piyush Harsh,
 *     URL: piyush-harsh.info
 */
 
package main 

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"io"
	"log"
	"os"
)

type user_struct struct {
    Username string `json:"username"`
    Password string `json:"password"`
    AdminFlag string `json:"isadmin"`
}

var (
	Trace	*log.Logger
	Info	*log.Logger
	Warning	*log.Logger
	Error	*log.Logger
	MyFileTrace	*log.Logger
	MyFileInfo	*log.Logger
	MyFileWarning	*log.Logger
	MyFileError	*log.Logger
	staticMsgs [20]string
)

func main() {
	file, err := os.OpenFile("auth-utils.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
    	log.Fatalln("Failed to open log file", "auth-utils.log", ":", err)
	}
	multi := io.MultiWriter(file, ioutil.Discard)
	Initlogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr, multi)
	//logger has been initialized at this point 
	InitMsgs()
	dbCheck := CheckDB("file:foo.db?cache=shared&mode=rwc")

	if dbCheck {
		MyFileInfo.Println("Table already exists in DB, nothing to do, proceeding normally.")
	} else {
		InitDB("file:foo.db?cache=shared&mode=rwc")
	}
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", HomeHandler)
	users := r.Path("/admin/user/").Subrouter()
	users.Methods("GET").HandlerFunc(UserListHandler)
	users.Methods("POST").HandlerFunc(UserCreateHandler)

    user := r.Path("/admin/user/{id}").Subrouter()
    user.Methods("GET").HandlerFunc(UserDetailsHandler)
    user.Methods("PUT").HandlerFunc(UserUpdateHandler)
    user.Methods("DELETE").HandlerFunc(UserDeleteHandler)

    auth := r.Path("/auth/{id}").Subrouter()
    auth.Methods("GET").HandlerFunc(UserAuthHandler)
	
	tokens := r.Path("/token/").Subrouter()
    tokens.Methods("POST").HandlerFunc(TokenGenHandler)
	
	tokenval := r.Path("/token/validate/{id}").Subrouter()
    tokenval.Methods("GET").HandlerFunc(TokenValidateHandler)

	MyFileInfo.Println("Starting server on :8000")
    http.ListenAndServe(":8000", r)
    MyFileInfo.Println("Stopping server on :8000")
}

func HomeHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	out.WriteHeader(http.StatusOK) //200 status code
	var jsonbody = staticMsgs[0]
    fmt.Fprintln(out, jsonbody)
    MyFileInfo.Println("Received request on URI:/ GET")
}