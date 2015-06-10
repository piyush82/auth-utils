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
 * Author: Piyush Harsh,
 * URL: piyush-harsh.info
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
	"io/ioutil"
	"io"
	"log"
	"os"
	"crypto/sha1"
	"strconv"
	"bytes"
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
	staticMsgs [10]string
)

func Initlogger(traceHandle, infoHandle, warningHandle, errorHandle, fileHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	MyFileTrace = log.New(fileHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	MyFileInfo = log.New(fileHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	MyFileWarning = log.New(fileHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	MyFileError = log.New(fileHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

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

	MyFileInfo.Println("Starting server on :8000")
    http.ListenAndServe(":8000", r)
    MyFileInfo.Println("Stopping server on :8000")
}

func UserListHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	userList := GetUserList("file:foo.db?cache=shared&mode=rwc", "user", "username")
	var jsonbody = staticMsgs[4]
	var buffer bytes.Buffer
	for i := 0; i < len(userList); i++ {
		if i == 0 {
			buffer.WriteString("\"");
			buffer.WriteString(userList[i])
			buffer.WriteString("\"")
		} else {
			buffer.WriteString(",\"");
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

func UserDetailsHandler(out http.ResponseWriter, in *http.Request) {
	id := mux.Vars(in)["id"]
	out.Header().Set("Content-Type", "application/json")
    fmt.Println("Showing details for user", id)
}

func UserAuthHandler(out http.ResponseWriter, in *http.Request) {
	id := mux.Vars(in)["id"]
	var passWord string
	out.Header().Set("Content-Type", "application/json")

	if len(in.Header["X-Auth-Password"]) == 0 {
		MyFileWarning.Println("Authentication Module - Can't Proceed: Password Missing! User =", id)
		out.WriteHeader(http.StatusBadRequest) //400 status code
		var jsonbody = staticMsgs[5]
		fmt.Fprintln(out, jsonbody)
	} else {
		passWord = in.Header["X-Auth-Password"][0]
		MyFileInfo.Println("A valid password [password hidden] received for user:", id)
		data := []byte(passWord)
    	hash := sha1.Sum(data)
    	sha1hash := string(hash[:])
    	MyFileInfo.Println("SHA-1 Hash Generated for the incoming password:", sha1hash)
    	//get the stored SHA-1 Hash for the incoming user
    	storedHash := LocatePasswordHash("file:foo.db?cache=shared&mode=rwc", "user", id)
    	MyFileInfo.Println("SHA-1 hash retrieved for the incoming user:", storedHash)
    	if strings.HasPrefix(storedHash, sha1hash) && strings.HasSuffix(storedHash, sha1hash) {
    		out.WriteHeader(http.StatusAccepted) //202 status code
    		var jsonbody = staticMsgs[7]
			fmt.Fprintln(out, jsonbody)
    		MyFileInfo.Println("Password matches. User", id, "successfully authenticated.")
    	} else {
    		out.WriteHeader(http.StatusUnauthorized) //401 status code
    		var jsonbody = staticMsgs[6]
			fmt.Fprintln(out, jsonbody)
    		MyFileInfo.Println("Password does not matche. User", id, "not authenticated.")
    	}
	}
	
    MyFileInfo.Println("Received request on URI:/auth/{user-id} GET")
}

func UserUpdateHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
}

func UserDeleteHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
}

func HomeHandler(out http.ResponseWriter, in *http.Request) {
	out.Header().Set("Content-Type", "application/json")
	out.WriteHeader(http.StatusOK) //200 status code
	var jsonbody = staticMsgs[0]
    fmt.Fprintln(out, jsonbody)
    MyFileInfo.Println("Received request on URI:/ GET")
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

func GetCount(filePath string, tableName string, columnName string, searchTerm string) int {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

    queryStmt := "SELECT count(*) FROM tablename WHERE columnname='searchterm';"
    queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
    queryStmt = strings.Replace(queryStmt, "columnname", columnName, 1)
    queryStmt = strings.Replace(queryStmt, "searchterm", searchTerm, 1)

    MyFileInfo.Println("SQLite3 Query:", queryStmt)

    rows, err := db.Query(queryStmt)
    if err != nil {
    	MyFileWarning.Println("Caught error in count method.")
    	checkErr(err, 1, db)
    }
    defer rows.Close()
    if rows.Next() {
    	var userCount int
        err = rows.Scan(&userCount)
        checkErr(err, 1, db)
        return userCount
    }
	return 1
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

func LocatePasswordHash(filePath string, tableName string, userName string) string {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

	queryStmt := "SELECT password FROM tablename WHERE username='searchterm';"
    queryStmt = strings.Replace(queryStmt, "tablename", tableName, 1)
    queryStmt = strings.Replace(queryStmt, "searchterm", userName, 1)
    
    MyFileInfo.Println("SQLite3 Query:", queryStmt)

	rows, err := db.Query(queryStmt)
    if err != nil {
    	MyFileWarning.Println("Caught error in user-password-locate method.")
    	checkErr(err, 1, db)
    }
    defer rows.Close()
    if rows.Next() {
    	var passWord string
        err = rows.Scan(&passWord)
        checkErr(err, 1, db)
        return passWord
    }
    
	return ""
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

func CheckDB(filePath string) bool {
	db, err := sql.Open("sqlite3", filePath)
	var status bool
	if err != nil {
        checkErr(err, 1, db)
    }
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	}

    rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
    if err != nil {
    	checkErr(err, 0, db)
    } else {
    	defer rows.Close()
    	for rows.Next() {
        	var tablename string
        	err = rows.Scan(&tablename)
        	checkErr(err, 1, db)
        	MyFileInfo.Println("While performing DB sanity checks: found table", tablename)
        	if len("user") == len(tablename) {
        		if strings.Count(tablename, "user") == 1 {
        			status = true
        		}
        	}
    	}
    }
    return status
}

func InitDB(filePath string) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
        checkErr(err, 1, db)
    } 
    defer db.Close()
    
    err = db.Ping()
	if err != nil {
    	panic(err.Error()) // proper error handling instead of panic in your app
	} else {

    	var dbCmd = `
    			CREATE TABLE 'user' (
    			'uid' INTEGER PRIMARY KEY AUTOINCREMENT,
    			'username' VARCHAR(64) NULL,
    			'password' VARCHAR(64) NULL,
    			'isadmin' VARCHAR(1)
				);
    		`
    	stmt, err := db.Prepare(dbCmd)
    	checkErr(err, 1, db)
    	res, err := stmt.Exec()
    	checkErr(err, 1, db)
    	dbCmd = `
    			CREATE TABLE 'token' (
    			'uid' INTEGER PRIMARY KEY AUTOINCREMENT,
    			'username' VARCHAR(64) NULL,
    			'validupto' VARCHAR(64) NULL,
    			'capability-list' VARCHAR(128)
				);
			`
		stmt, err = db.Prepare(dbCmd)
    	checkErr(err, 1, db)
    	res, err = stmt.Exec()
    	checkErr(err, 1, db)
    	MyFileInfo.Println("Created tables user, token. System ready. System response:", res)
    }
}

func checkErr(err error, errorType int, db *sql.DB) {
    if err != nil {
    	MyFileError.Println("Unrecoverable Error!", err)
    	panic(err)
    }
}

func InitMsgs() {
	staticMsgs[0] = 
`[
	{
		"metadata": 
		{
			"source": "T-Nova-AuthZ-Service"
		},
		"info":
		[
			{
				"msg": "Welcome to the T-Nova-AuthZ-Service",
				"purpose": "REST API Usage Guide",
				"disclaimer": "It's not yet final!",
				"notice": "Headers and body formats are not defined here."
			}
		],
		"api":
		[
			{
				"uri": "/",
				"method": "GET",
				"purpose": "REST API Structure and Capability Discovery"
			},
			{
				"uri": "/admin/user/",
				"method": "GET",
				"purpose": "Admin API to get list of all users"
			},
			{
				"uri": "/admin/user/",
				"method": "POST",
				"purpose": "Create a new user"
			},
			{
				"uri": "/admin/user/{user-id}",
				"method": "GET",
				"purpose": "Admin API to get detailed info of a particular user"
			},
			{
				"uri": "/admin/user/{user-id}",
				"method": "PUT",
				"purpose": "Admin API to modify details of a particular user"
			},
			{
				"uri": "/admin/user/{user-id}",
				"method": "DELETE",
				"purpose": "Admin API to delete a particular user"
			},
			{
				"uri": "/token/",
				"method": "POST",
				"purpose": "API to request a new service token"
			},
			{
				"uri": "/token/{token-uuid}",
				"method": "GET",
				"purpose": "Get details of this token, lifetime, user or service id of the creator, etc."
			},
			{
				"uri": "/token/{token-uuid}",
				"method": "DELETE",
				"purpose": "Revoke or essentially delete an existing token."
			},
			{
				"uri": "/token/{token-uuid}/validate",
				"method": "GET",
				"purpose": "Validate a existing token, OK/Not-OK type status response"
			},
			{
				"uri": "/token/{token-uuid}/capability",
				"method": "GET",
				"purpose": "Get the capabilities associated with this token"
			},
			{
				"uri": "/auth/{user-id}",
				"method": "GET",
				"purpose": "Authenticate an existing user. OK/Not-OK type response."
			}
		]
	}
]`
	staticMsgs[1] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "No or corrupt POST data received."
		}
	]
}`
	staticMsgs[2] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "User already exists.",
		}
	]
}`
	staticMsgs[3] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "user created susseccfully",
			"auth-uri": "/auth/xxx",
			"admin-uri": "/admin/user/yyy"
		}
	]
}`
	staticMsgs[4] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "list of active users",
		}
	],
	"userlist":
	[
		xxx
	]
}`
	staticMsgs[5] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "Incorrect / Missing Header Attributes."
		}
	]
}`
	staticMsgs[6] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "Incorrect Password."
		}
	]
}`
	staticMsgs[7] = 
`
{
	"metadata": 
	{
		"source": "T-Nova-AuthZ-Service"
	},
	"info":
	[
		{
			"msg": "Authentication Successful."
		}
	]
}`
}