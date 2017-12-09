package main

import (
    // "log"
    // "time"

    "io/ioutil"
    "encoding/json"

    "net/http"
    "crypto/tls"
)

type AccessToken struct {
  Token string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  TokenType string `json:"token_type"`
  ExpiresIn int `json:"expires_in"`
  scope string `json:"scope"`
}

func getAuthBearer(aokURL string) (AccessToken, error) {
  tr := &http.Transport{
	   TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
  }
  client := &http.Client{Transport: tr}

  var token AccessToken

  req, err := http.NewRequest("POST", aokURL, nil)
  req.Header.Add("Authorization", "Basic Y2Y6")
  resp, err := client.Do(req)
  if err != nil {
    return token, err
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return token, err
  }
  resp.Body.Close()

  // log.Println(string(body))

  err = json.Unmarshal(body, &token)
  if err != nil {
    return token, err
  }

  return token, nil

}



// func main () {
//   authBearer, err := getAuthBearer(stackatoAOKURL)
//   if err != nil {
//     log.Fatal(err)
//   }
//   tokenExpireAt := time.Now().Add(time.Duration(authBearer.ExpiresIn - 240)*time.Second)
//
//   log.Printf("Auth: %+v\n", authBearer)
//   log.Println(authBearer.ExpiresIn)
//   log.Println(time.Now())
//   log.Println(tokenExpireAt)
//   log.Println(tokenExpireAt.Before(time.Now()))
// }
