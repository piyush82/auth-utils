# Gatekeeper
Welcome to the *Gatekeeper* module istallation guide. This package is written in Go programming language.

## Install Go runtime
Please follow the download and install instrictions located here: https://golang.org/doc/install. Also you need to prepare your environment for optimal Go experience. Please read and follow the helpful instructions found here: https://golang.org/doc/code.html

## Installing Gatekeeper
After you have cloned the source code in the proper source directory structure as specified in the guides mentioned above, installing *Gatekeeper* is very easy. Simply follow these steps -

1. From inside the source folder where you copied the *Gatekeeper* files, simply run `go get`
2. Then run `go install`
3. If you followed the environment setup instructions, you should be able to launch *Gatekeeper* by simply typing `auth-utils`from any place.
4. Alternatively, from within the source folder where *Gatekeeper* code files were copied into, do `go run *.go` to launch the m-service.

```
The service will start at port 8000
```


## API and Usage Guide
Please see the T-Nova internal wiki page for API example snippets, [T-Nova Gatekeeper](http://wiki.t-nova.eu/tnovawiki/index.php/Gatekeeper)


## Word of caution
This is a v0 release, the program keeps the database files and stores the log files in the folder from where the program was executed, regardless of where the program binary is. So in order to avoid any suprises, execute *Gatekeeper* program from the same path always. 

```
This path issue will be resolve in a future release. Also the port to start the service will be made configurable.
```

## Credits
### Development Team
* Piyush Harsh (harh@zhaw.ch) / ICCLab