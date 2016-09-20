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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func DcListHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	if len(in.Header["X-Auth-Token"]) == 0 {
		MyFileWarning.Println("User List Module - Can't Proceed: Token Missing!")
		out.WriteHeader(http.StatusBadRequest) //400 status code
		var jsonbody = staticMsgs[5]
		fmt.Fprintln(out, jsonbody)
	} else {
		token := in.Header["X-Auth-Token"][0]
		//check if token is valid and belongs to an admin user
		isAdmin := CheckTokenAdmin(token)
		if isAdmin {
			dcNameList, dcIDList := GetDcList(dbArg, "dcdata", "dcname", "did")
			var jsonbody = staticMsgs[19]
			var buffer bytes.Buffer
			var buffer2 bytes.Buffer
			for i := 0; i < len(dcNameList); i++ {
				if i == 0 {
					buffer.WriteString("\"")
					buffer.WriteString(dcNameList[i])
					buffer.WriteString("\"")
					buffer2.WriteString("\"")
					buffer2.WriteString(dcIDList[i])
					buffer2.WriteString("\"")
				} else {
					buffer.WriteString(",\"")
					buffer.WriteString(dcNameList[i])
					buffer.WriteString("\"")
					buffer2.WriteString(",\"")
					buffer2.WriteString(dcIDList[i])
					buffer2.WriteString("\"")
				}
			}
			jsonbody = strings.Replace(jsonbody, "xxx", buffer.String(), 1)
			jsonbody = strings.Replace(jsonbody, "yyy", buffer2.String(), 1)
			out.WriteHeader(http.StatusOK) //200 status code
			fmt.Fprintln(out, jsonbody)
		} else {
			var jsonbody = staticMsgs[18]
			out.WriteHeader(http.StatusUnauthorized) //401 status code
			fmt.Fprintln(out, jsonbody)
		}
	}

	MyFileInfo.Println("Received request on URI:/admin/dc/ GET")
}

func DcCreateHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	if len(in.Header["X-Auth-Token"]) == 0 {
		MyFileWarning.Println("DC Create Module - Can't Proceed: X-Auth-Token Missing!")
		out.WriteHeader(http.StatusBadRequest) //400 status code
		var jsonbody = staticMsgs[5]
		fmt.Fprintln(out, jsonbody)
	} else {
		token := in.Header["X-Auth-Token"][0]
		//check if token is valid and belongs to an admin user
		isAdmin := CheckTokenAdmin(token)
		if isAdmin {
			decoder := json.NewDecoder(in.Body)
			var u dc_struct
			err := decoder.Decode(&u)

			if err != nil {
				out.WriteHeader(http.StatusBadRequest) //status 400 Bad Request
				var jsonbody = staticMsgs[1]
				fmt.Fprintln(out, jsonbody)
				MyFileInfo.Println("Received malformed request on URI:/admin/dc/ POST")
				panic(err)
			} else if len(u.Dcname) == 0 || len(u.AdminId) == 0 || len(u.Password) == 0 {
				MyFileInfo.Println("Received malformed request on URI:/admin/dc/ POST")
				out.WriteHeader(http.StatusBadRequest)
				var jsonbody = staticMsgs[1] //status 400 Bad Request
				fmt.Fprintln(out, jsonbody)
			} else {
				MyFileInfo.Println("Received JSON: Struct value received for dc [pass hidden]:", u.Dcname, "with admin-id:", u.AdminId)
				dcCount := GetCount(dbArg, "dcdata", "dcname", u.Dcname)
				if dcCount > 0 {
					MyFileInfo.Println("Duplicate dc create request on URI:/admin/dc/ POST")
					out.WriteHeader(http.StatusPreconditionFailed)
					var jsonbody = staticMsgs[20] //datacenter already exists
					fmt.Fprintln(out, jsonbody)
				} else {
					//now store the new datacenter in the table and return back the proper response
					MyFileInfo.Println("Attempting to store new dc:", u.Dcname, "into the table.")
					status := InsertDc(dbArg, "dcdata", u.Dcname, u.AdminId, u.Password, u.ExtraInfo) //inserting extra info now
					MyFileInfo.Println("Status of the attempt to store new datacenter:", u.Dcname, "into the table was:", status)

					out.WriteHeader(http.StatusOK) //200 status code
					var jsonbody = staticMsgs[21]  //user user creation msg, replace with actual content for xxx and yyy
					dcId := LocateDc(dbArg, "dcdata", u.Dcname)
					MyFileInfo.Println("The new id for datacenter:", u.Dcname, "is:", dcId)
					//constructing the correct JSON response
					jsonbody = strings.Replace(jsonbody, "yyy", strconv.Itoa(dcId), 1)
					jsonbody = strings.Replace(jsonbody, "zzz", strconv.Itoa(dcId), 1)
					fmt.Fprintln(out, jsonbody)
				}
			}
		} else {
			var jsonbody = staticMsgs[18]
			out.WriteHeader(http.StatusUnauthorized) //401 status code
			fmt.Fprintln(out, jsonbody)
		}
	}
	MyFileInfo.Println("Received request on URI:/admin/dc/ POST")
}

func DcDetailsHandler(out http.ResponseWriter, in *http.Request) {
	id := mux.Vars(in)["id"]
	out.Header().Set("Content-Type", "application/json")
	if len(in.Header["X-Auth-Token"]) == 0 {
		MyFileWarning.Println("DC Details Module - Can't Proceed: Token Missing!")
		out.WriteHeader(http.StatusBadRequest) //400 status code
		var jsonbody = staticMsgs[5]
		fmt.Fprintln(out, jsonbody)
	} else {
		token := in.Header["X-Auth-Token"][0]
		//check if token is valid and belongs to an admin user
		isAdmin := CheckTokenAdmin(token)
		if isAdmin {
			dcDetail := GetDcDetail(dbArg, "dcdata", id)
			if dcDetail != nil {
				var jsonbody = staticMsgs[22]
				jsonbody = strings.Replace(jsonbody, "xxx", dcDetail[0], 1)
				jsonbody = strings.Replace(jsonbody, "yyy", dcDetail[1], 1)
				jsonbody = strings.Replace(jsonbody, "zzz", dcDetail[2], 1)
				jsonbody = strings.Replace(jsonbody, "aaa", dcDetail[3], 1)
				out.WriteHeader(http.StatusOK) //200 status code
				fmt.Fprintln(out, jsonbody)
			} else {
				out.WriteHeader(http.StatusNotFound) //404 status code
				var jsonbody = staticMsgs[15]
				fmt.Fprintln(out, jsonbody)
			}
		} else {
			var jsonbody = staticMsgs[18]
			out.WriteHeader(http.StatusUnauthorized) //401 status code
			fmt.Fprintln(out, jsonbody)
		}
	}
	MyFileInfo.Println("Received request on URI:/admin/user/{id} GET for uid:", id)
}

func DcUpdateHandler(out http.ResponseWriter, in *http.Request) {

}

func DcDeleteHandler(out http.ResponseWriter, in *http.Request) {
	id := mux.Vars(in)["id"]
	out.Header().Set("Content-Type", "application/json")
	if len(in.Header["X-Auth-Token"]) == 0 {
		MyFileWarning.Println("DC Delete Module - Can't Proceed: Token Missing!")
		out.WriteHeader(http.StatusBadRequest) //400 status code
		var jsonbody = staticMsgs[5]
		fmt.Fprintln(out, jsonbody)
	} else {
		token := in.Header["X-Auth-Token"][0]
		//check if token is valid and belongs to an admin user
		isAdmin := CheckTokenAdmin(token)
		if isAdmin {
			//now delete this dc from the table
			MyFileInfo.Println("Attempting to delete dc:", id, "from the table.")
			status := DeleteDc(dbArg, "dcdata", id)
			MyFileInfo.Println("Status of the attempt to delete existing dcid:", id, "from the table was:", status)
			if status == 1 {
				out.WriteHeader(http.StatusOK) //200 status code
				var jsonbody = staticMsgs[25]  //dc deletion msg
				fmt.Fprintln(out, jsonbody)
			} else {
				out.WriteHeader(http.StatusInternalServerError) //500 status code
				var jsonbody = staticMsgs[26]
				fmt.Fprintln(out, jsonbody) //dc deletion failed msg
			}
		} else {
			var jsonbody = staticMsgs[18]
			out.WriteHeader(http.StatusUnauthorized) //401 status code
			fmt.Fprintln(out, jsonbody)
			MyFileInfo.Println("Received unauthorized request on URI:/admin/dc/{id} PUT for dcid:", id)
		}
	}
	MyFileInfo.Println("Received request on URI:/admin/dc/{id} DELETE")
}

func GetDcList(filePath string, tableName string, columnName1 string, columnName2 string) ([]string, []string) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		checkErr(err, 1, db)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT column1, column2 FROM tablename;"
	queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
	queryStmt = strings.Replace(queryStmt, "column1", columnName1, 1)
	queryStmt = strings.Replace(queryStmt, "column2", columnName2, 1)

	MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
	if err != nil {
		MyFileWarning.Println("Caught error in dc-list method.")
		checkErr(err, 1, db)
	}
	defer rows.Close()
	var dclist []string
	var dcid []string
	for rows.Next() {
		var dcName string
		var dcID string
		err = rows.Scan(&dcName, &dcID)
		checkErr(err, 1, db)
		dclist = append(dclist, dcName)
		dcid = append(dcid, dcID)
	}
	return dclist, dcid
}

func InsertDc(filePath string, tableName string, dcName string, dcAdmin string, password string, xtraInfo string) bool {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		checkErr(err, 1, db)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	insertStmt := "INSERT INTO tablename VALUES (NULL, 'dcname', 'dcadmin', 'password', 'xtra');"
	insertStmt = strings.Replace(insertStmt, "tablename", tableName, 1)
	insertStmt = strings.Replace(insertStmt, "dcname", dcName, 1)
	insertStmt = strings.Replace(insertStmt, "xtra", xtraInfo, 1)
	insertStmt = strings.Replace(insertStmt, "password", password, 1)
	insertStmt = strings.Replace(insertStmt, "dcadmin", dcAdmin, 1)
	MyFileInfo.Println("SQLite3 Query:", insertStmt)

	res, err := db.Exec(insertStmt)
	if err != nil {
		MyFileWarning.Println("Caught error in insert-datacenter method,", res)
		checkErr(err, 1, db)
	}

	return true
}

func DeleteDc(filePath string, tableName string, dcId string) int {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		checkErr(err, 1, db)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "DELETE FROM tablename WHERE did=val;"
	queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
	queryStmt = strings.Replace(queryStmt, "val", dcId, 1)

	MyFileInfo.Println("SQLite3 Query:", queryStmt)

	_, err = db.Exec(queryStmt)
	if err != nil {
		MyFileWarning.Println("Caught error in dc-delete method.")
		checkErr(err, 1, db)
		return 0
	}
	return 1
}

func GetDcDetail(filePath string, tableName string, dcId string) []string {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		checkErr(err, 1, db)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT dcname, adminid, password, xtrainfo FROM tablename WHERE did=val;"
	queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
	queryStmt = strings.Replace(queryStmt, "val", dcId, 1)

	MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
	if err != nil {
		MyFileWarning.Println("Caught error in datacenter-detail method.")
		checkErr(err, 1, db)
	}
	defer rows.Close()
	var dcdetail []string
	if rows.Next() {
		var dcName string
		var adminId string
		var passWord string
		var xtraInfo string
		err = rows.Scan(&dcName, &adminId, &passWord, &xtraInfo)
		checkErr(err, 1, db)
		dcdetail = append(dcdetail, dcName)
		dcdetail = append(dcdetail, adminId)
		dcdetail = append(dcdetail, passWord)
		dcdetail = append(dcdetail, xtraInfo)
	} else {
		dcdetail = nil
	}
	return dcdetail
}

func LocateDc(filePath string, tableName string, dcName string) int {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		checkErr(err, 1, db)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT did FROM tablename WHERE dcname='searchterm';"
	queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
	queryStmt = strings.Replace(queryStmt, "searchterm", dcName, 1)

	MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
	if err != nil {
		MyFileWarning.Println("Caught error in dc-locate method.")
		checkErr(err, 1, db)
	}
	defer rows.Close()
	if rows.Next() {
		var dcId int
		err = rows.Scan(&dcId)
		checkErr(err, 1, db)
		return dcId
	}

	return -1
}
