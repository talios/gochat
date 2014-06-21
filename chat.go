package main

import (
  "fmt"
  "code.google.com/p/go-uuid/uuid"
  "os"
  "flag"
  "net"
  "bufio"
  "strings"
)

type Message struct {
  id string
  message string
  final bool
}

var seen map[string]bool
var mirrors map[string]net.Conn

func getServerName() (name string) {
  fmt.Printf("Server Name:!")

  _, err := fmt.Scanf("%s", &name)

  if (nil != err) {
    panic("Cannot determine server name")
  }

  fmt.Println("Server now named ", name)

  return
}

func interactWithUser(serverName string, messageChannel chan Message) {
  var finished bool = false
  var line string
  for !finished {
    fmt.Print(">")
    fmt.Scanf("%s", &line)
    finished = line == "."
    if (finished) {
      line = fmt.Sprintf("%s has left the conversation.", serverName)
      delete(mirrors, serverName)
    }

    msg := Message{id:uuid.NewUUID().String(), message: line, final: finished}

    messageChannel <- msg
  }
}

/*
 Go-Routine
 */
func handleIncomingRequest(conn net.Conn, messageChannel chan Message) {
  reader := bufio.NewReader(conn)

  for {
    line, err := reader.ReadString(byte('\n'))

    if (err != nil ) {
      fmt.Println("Network connection died.")
      break
    }

    firstSpace := strings.Index(line, " ")
    if firstSpace != -1 {
      msg := Message{id:line[0:firstSpace], message: line[firstSpace+1:len(line)-1]}
      messageChannel <- msg
    }
  }
}

func listenForMessages(port string, messageChannel chan Message) {
  ln, err := net.Listen("tcp", ":" + port)
  if (err != nil) {
    panic("Can't listen to network port")
  }
  for {
    conn, err := ln.Accept()
    if (err == nil) {
      fmt.Println("Received connection")
      go handleIncomingRequest(conn, messageChannel)
      fmt.Printf("Connecting back to %s\n", conn.RemoteAddr().String())
      mirrors[conn.RemoteAddr().String()] = conn
    } else {
      fmt.Println("Failed receiving incoming request")
    }
  }
}

func connectToMirror(mirror string, messageChannel chan Message) net.Conn {
  if (mirror != "") {
    conn, err := net.Dial("tcp", mirror)
    if (err != nil) {
      panic("Can't connect to mirror")
    }
    mirrors[mirror] = conn
    go handleIncomingRequest(conn, messageChannel)
    return conn
  }
  return nil

}

func main() {
  seen = make(map[string]bool)
  mirrors = make(map[string]net.Conn)
  var mirror string
  var port string

  flag.StringVar(&port, "port", "7030", "listen port")
  flag.StringVar(&mirror, "mirror", "", "mirror server")
  flag.Parse()

  serverName := getServerName()
  messageChannel := make(chan Message)

  go listenForMessages(port, messageChannel)

  connectToMirror(mirror, messageChannel)

  go interactWithUser(serverName, messageChannel)

  for count := 1; count < len(os.Args); count ++ {
    fmt.Println("connecting to ", os.Args[count])
  }

  for msg := range messageChannel {
    if !seen[msg.id] {
      fmt.Printf("Echo> %s: %s\n", msg.id, msg.message)

      seen[msg.id] = true

      for _, value := range mirrors {
          fmt.Fprintf(value, "%s %s\n", msg.id, msg.message)
      }

    }

    if (msg.final) {
      break
    }
  }
}
