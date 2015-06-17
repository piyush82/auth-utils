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
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"encoding/json"
	"crypto/sha1"
	"strconv"
	"bytes"
)

func UserDetailsHandler(out http.ResponseWriter, in *http.Request) {
	id := mux.Vars(in)["id"]
	out.Header().Set("Content-Type", "application/json")
    fmt.Println("Showing details for user", id)
}

func UserUpdateHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
}

func UserDeleteHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
}

func UserListHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	userList := GetUserList("file:foo.db?cache=shared&mode=rwc", "user", "username")
	var jsonbody = staticMsgs[4]
	var buffer bytes.Buffer
	for i := 0; i < len(userList); i++ {
		if i == 0 {
			buffer.WriteString("\"")
			buffer.WriteString(userList[i])
			buffer.WriteString("\"")
		} else {
			buffer.WriteString(",\"")
			buffer.WriteString(userList[i])
			buffer.WriteString("\"")
		}
	}
	jsonbody = strings.Replace(jsonbody, "xxx", buffer.String(), 1)
	out.WriteHeader(http.StatusOK) //200 status code
	fmt.Fprintln(out, jsonbody)
	MyFileInfo.Println("Received request on URI:/admin/user/ GET")
}

func UserCreateHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(in.Body)
	var u user_struct   
    err := decoder.Decode(&u)

    if err != nil {
    	out.WriteHeader(http.StatusBadRequest) //status 400 Bad Request
    	var jsonbody = staticMsgs[1]
		fmt.Fprintln(out, jsonbody)
		MyFileInfo.Println("Received malformed request on URI:/admin/user/ POST")
        panic(err)
    } else if len(u.Username) == 0 {
    	MyFileInfo.Println("Received malformed request on URI:/admin/user/ POST")
    	out.WriteHeader(http.StatusBadRequest)
    	var jsonbody = staticMsgs[1] //status 400 Bad Request
		fmt.Fprintln(out, jsonbody)
    } else {
    	MyFileInfo.Println("Received JSON: Struct value received for user [pass hidden]:", u.Username)
    	userCount := GetCount("file:foo.db?cache=shared&mode=rwc", "user", "username", u.Username)
    	if userCount > 0 {
    		MyFileInfo.Println("Duplicate user create request on URI:/admin/user/ POST")
    		out.WriteHeader(http.StatusPreconditionFailed)
    		var jsonbody = staticMsgs[2] //user already exists
			fmt.Fprintln(out, jsonbody)
    	} else {
    		//now store the new user in the table and return back the proper response
    		MyFileInfo.Println("Attempting to store new user:", u.Username, "into the table.")
    		status := InsertUser("file:foo.db?cache=shared&mode=rwc", "user", u.Username, u.Password, u.AdminFlag)
    		MyFileInfo.Println("Status of the attempt to store new user:", u.Username, "into the table was:", status)

    		out.WriteHeader(http.StatusOK) //200 status code
    		var jsonbody = staticMsgs[3] //user user creation msg, replace with actual content for xxx and yyy
    		uId := LocateUser("file:foo.db?cache=shared&mode=rwc", "user", u.Username)
    		MyFileInfo.Println("The new id for user:", u.Username, "is:", uId)
    		//constructing the correct JSON response
    		jsonbody = strings.Replace(jsonbody, "xxx", strconv.Itoa(uId), 1)
    		jsonbody = strings.Replace(jsonbody, "yyy", strconv.Itoa(uId), 1)
			fmt.Fprintln(out, jsonbody)
    	}
    	
		MyFileInfo.Println("Received request on URI:/admin/user/ POST")
    }
}

func GetUserList(filePath string, tableName string, columnName string) []string {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT column FROM tablename;"
	queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
	queryStmt = strings.Replace(queryStmt, "column", columnName, 1)

	MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
    if err != nil {
    	MyFileWarning.Println("Caught error in user-list method.")
    	checkErr(err, 1, db)
    }
    defer rows.Close()
    var ulist []string
    for rows.Next() {
    	var userName string
        err = rows.Scan(&userName)
        checkErr(err, 1, db)
        ulist = append(ulist, userName)
    }
    return ulist
}

func LocateUser(filePath string, tableName string, userName string) int {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT uid FROM tablename WHERE username='searchterm';"
    queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
    queryStmt = strings.Replace(queryStmt, "searchterm", userName, 1)
    
    MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
    if err != nil {
    	MyFileWarning.Println("Caught error in user-locate method.")
    	checkErr(err, 1, db)
    }
    defer rows.Close()
    if rows.Next() {
    	var userId int
        err = rows.Scan(&userId)
        checkErr(err, 1, db)
        return userId
    }
    
	return -1
}

func InsertUser(filePath string, tableName string, userName string, passWord string, isAdmin string) bool {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

    insertStmt := "INSERT INTO tablename VALUES (NULL, 'username', 'passhash', 'isadmin');"
    insertStmt = strings.Replace(insertStmt, "tablename", tableName, 1)
    insertStmt = strings.Replace(insertStmt, "username", userName, 1)
    data := []byte(passWord)
    hash := sha1.Sum(data)
    sha1hash := string(hash[:])
    MyFileInfo.Println("SHA-1 Hash Generated for the incoming password:", sha1hash)

    insertStmt = strings.Replace(insertStmt, "passhash", sha1hash, 1)
    insertStmt = strings.Replace(insertStmt, "isadmin", isAdmin, 1)
    MyFileInfo.Println("SQLite3 Query:", insertStmt)

    res, err := db.Exec(insertStmt)
    if err != nil {
    	MyFileWarning.Println("Caught error in insert-user method,", res)
    	checkErr(err, 1, db)
    }
    
	return true
}