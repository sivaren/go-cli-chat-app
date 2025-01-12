# Go Chat CLI-based Application
Go Chat CLI-based Application

> _Implemented chat application using Gorilla Websocket._

## Table of Contents
* [General Information](#general-information)
* [Requirements](#requirements)
* [Prerequisites](#prerequisites)
* [Run Program](#run-program)
    * [Using Command-line](#using-command-line)
    * [Using Docker](#using-docker)
* [Project Status](#project-status)
* [Functionality](#functionality)

## General Information 
> A CLI-based chat application with Go 

## Requirements 
* Go
* github.com/gorilla/websocket
* golang.org/x/crypto

## Prerequisites
> **Ensure that you're in the `main` branch** </br>

**Clone this repository using the following command line (git bash)**
```
$ git clone https://github.com/sivaren/go-cli-chat-app.git 
```

## Run Program

### Using Command-line
* Open `cmd` on this folder and download dependencies

  ```
  go mod tidy
  go mod download
  ```
* Run websocket `server`
  ```
  go run ./server
  ```
* Open new terminal to run a `client`
  ```
  go run ./client
  ```

### Using Docker
* Open `cmd` on this folder and `build` Docker image

  ```
  docker build -t go-cli-chat-app .
  ```
* Run websocket `server`
  ```
  docker run -p 8080:8080 go-cli-chat-app
  ```
* Open new terminal to run a `client`
  ```
  go run ./client
  ```

## Project Status
> **Project is: DONE _(not fully implemented)_**

## Functionality
<table>
    <tr>
      <td><b>Use Case</b></td>
      <td><b>Status</b></td>
    </tr>
    <tr>
      <td>Broadcast Message</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>Direct Message</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>User Authentication</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>User Join and Leave Chat Room</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>Users Create Chat Rooms</td>
      <td>Not Implemented</td>
    </tr>
    <tr>
      <td>Store Messages</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>Error Handling</td>
      <td>Implemented</td>
    </tr>
    <tr>
      <td>Docker</td>
      <td>Implemented</td>
    </tr>
</table>
