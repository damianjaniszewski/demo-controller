package main

import (
    "fmt"
    "log"
    "os"
    "net/http"
    "encoding/json"
    "math/rand"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

var (
  queueName = os.Getenv("queueName")
  stackatoAOKURL = os.Getenv("stackatoAOKURL")
)

type restMessage struct {
  Value int `json:"value"`
}

func restHandler(w http.ResponseWriter, r *http.Request) {

    message := restMessage{rand.Intn(129)}
    output, err := json.Marshal(&message)

    if err != nil {
      log.Printf("update error: %s\n", err)
    } else {
      log.Printf("update: %s\n", string(output))
    }

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    fmt.Fprintln(w, string(output))
}

type wsSendMessage struct {
  ID int `json:"id"`
  Result restMessage `json:"result"`
}

type wsReadMessage struct {
  Op      string    `json:"op"`
  ID      int       `json:"id"`
  Url     string    `json:"url"`
  Params  map[string]string  `json:"params"`
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type wsRegister struct {
  ID        int
  Name      string
  Conn      *websocket.Conn
}

var sockets map[string]wsRegister

func wsHandler(w http.ResponseWriter, r *http.Request) {

    conn, err := upgrader.Upgrade(w, r, nil)

    if err != nil {
      log.Printf("websocket upgrade error: %s\n", err)
      return
    }
    for {
      messageType, p, err := conn.ReadMessage()
      if err != nil {
        log.Printf("websocket read error: %s\n", err)
        return
      }
      log.Printf("websocket received: %s [message type: %d]\n", p, messageType)

      var msgReceived wsReadMessage
      if err = json.Unmarshal(p, &msgReceived); err != nil {
        log.Printf("unmarshal error: %s\n", err)
        return
      }

      sockets[msgReceived.Params["uuid"]] = wsRegister{msgReceived.ID, msgReceived.Params["name"], conn}
      log.Printf("websocket registered: uuid:%s, name:%s, id:%d, conn:%s\n", msgReceived.Params["uuid"], msgReceived.Params["name"], msgReceived.ID, conn)

      message := wsSendMessage{msgReceived.ID, restMessage{0}}
      output, _ := json.Marshal(&message)
      log.Printf("websocket send: %s\n", output)
      if err = conn.WriteMessage(websocket.TextMessage, output); err != nil {
        log.Printf("websocket write error: %s\n", err)
        return
      }

    }
}

func main() {

  sockets = make(map[string]wsRegister)

  go func() {
    // authBearer, err := getAuthBearer(stackatoAOKURL)
    // if err != nil {
    //   log.Println(err)
    // }
    // log.Println("authBearer:", authBearer)
    // dashboard_url, username, password, err := getService(stackatoURL, serviceName, authBearer)
    // if err != nil {
    //   log.Println(err)
    // }
    // log.Println("RabbitMQ Dashboard:", dashboard_url, username, password)

    var (
      message = wsSendMessage{0, restMessage{0}}
      authBearer AccessToken
      tokenExpireAt = time.Now()
    )

    for {
      queueLen, err := getQueueLen(queueName)
      if err != nil {
        log.Println(err)
      }
      // log.Println("Queue len:", queueLen)

      // log.Println(tokenExpireAt, time.Now(), tokenExpireAt.Before(time.Now()))
      if tokenExpireAt.Before(time.Now()) {
        authBearer, err = getAuthBearer(stackatoAOKURL)
        if err != nil {
          log.Fatal(err)
        }
        tokenExpireAt = time.Now().Add(time.Duration(authBearer.ExpiresIn - 240)*time.Second)
        log.Println("Auth token:", authBearer.TokenType+" "+authBearer.Token, "\nExpires at:", tokenExpireAt)
      }

      instCount, cpuTotal, cpuAvg, err := getAppStats(os.Getenv("appName"), authBearer.TokenType+" "+authBearer.Token)
      if err != nil {
        log.Println(err)
      }
      log.Println("Queue len:", queueLen, "App stats:", instCount, cpuTotal, cpuAvg, len(sockets))

      for uuid, ws := range sockets {

        switch ws.Name {
          case "Queue": message = wsSendMessage{ws.ID, restMessage{queueLen}}
          case "Instances": message = wsSendMessage{ws.ID, restMessage{instCount}}
          case "CPU": message = wsSendMessage{ws.ID, restMessage{int(cpuTotal)}}
          case "CPUavg": message = wsSendMessage{ws.ID, restMessage{int(cpuAvg)}}
          case "Clients": message = wsSendMessage{ws.ID, restMessage{len(sockets)}}
          default: message = wsSendMessage{ws.ID, restMessage{0}}
        }
        output, err := json.Marshal(&message)
        if err != nil {
          log.Println(err)
        }
        err = ws.Conn.WriteMessage(websocket.TextMessage, output)
        if err != nil {
          log.Println(err)
          delete(sockets, uuid)
          log.Printf("websocket deleted: Name: %s, ID: %s\n", ws.Name, uuid)
        } else {
          // log.Printf("websocket send: %s (Name: %s, ID: %s)\n", output, ws.Name, uuid)
        }
      }
      time.Sleep(2 * time.Second)
    }
  }()

  router := mux.NewRouter().StrictSlash(true)
  router.HandleFunc("/ws", wsHandler)
  router.HandleFunc("/update", restHandler)

  log.Println("Demo Controller started, listening on port ", os.Getenv("PORT"))
  log.Println("> stackatoAOKURL: ", os.Getenv("stackatoAOKURL"))
  log.Println("> stackatoURL: ", os.Getenv("stackatoURL"))
  log.Println("> appName: ", os.Getenv("appName"))
  log.Println("> queueName: ", os.Getenv("queueName"))

  log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}
