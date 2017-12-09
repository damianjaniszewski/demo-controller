package main

import (
    // "log"
    "errors"
    "os"

    "io/ioutil"
    "encoding/json"

    "net/http"
    "crypto/tls"
)

var (
  stackatoURL = os.Getenv("stackatoURL")
  appName = os.Getenv("appName")
)

type Entity struct {
  Name string `json:"name"`
  Memory int `json:"memory"`
  Instances int `json:"instances"`
  DiskQuota int `json:"disk_quota"`
  State string `json:"state"`
}

type Metadata struct {
  GUID string `json:"guid"`
  URL string `json:"url"`
  created_at string `json:"created_at"`
  updated_at string `json:"updated_at"`
}

type Resource struct {
  Metadata Metadata `json:"metadata"`
  Entity Entity `json:"entity"`
}

type Apps struct {
  total_results int `json:"total_results"`
  total_pages int `json:"total_pages"`
  prev_url string `json:"prev_url"`
  next_url string `json:"next_url"`
  Resources []Resource `json:"resources"`
}

func getAppStats(appName string, authBearer string) (int, float64, float64, error) {
  tr := &http.Transport{
	   TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
  }
  client := &http.Client{Transport: tr}

  req, err := http.NewRequest("GET", stackatoURL + "/v2/apps", nil)
  if err != nil {
    return 0, 0, 0, err
  }

  // log.Println(authBearer)
  req.Header.Add("Authorization", authBearer)
  resp, err := client.Do(req)
  if err != nil {
    return 0, 0, 0, err
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0, 0, 0, err
  }
  resp.Body.Close()

  var f Apps
  err = json.Unmarshal(body, &f)
  if err != nil {
    return 0, 0, 0, err
  }

  appURL := ""
  instCount := 0
  for i := range f.Resources {
    if f.Resources[i].Entity.Name == appName && f.Resources[i].Entity.State == "STARTED" {
      appURL = f.Resources[i].Metadata.URL + "/stats"
      instCount = f.Resources[i].Entity.Instances
    }
  }

  if instCount == 0 {
    return 0, 0, 0, errors.New("Cannot get app stats for: " + appName)
  }

  req, err = http.NewRequest("GET", stackatoURL + appURL, nil)
  if err != nil {
    return 0, 0, 0, err
  }

  req.Header.Add("Authorization", authBearer)
  resp, err = client.Do(req)
  if err != nil {
    return 0, 0, 0, err
  }

  body, err = ioutil.ReadAll(resp.Body)
  if err != nil {
    return 0, 0, 0, err
  }
  resp.Body.Close()

  var g map[string]interface{}
  err = json.Unmarshal(body, &g)
  if err != nil {
    return 0, 0, 0, err
  }

  var cpuTotal float64 = 0.0
  var cpuAvg float64 = 0.0
  for k, _ := range g {
    if g[k].(map[string]interface{})["state"] == "RUNNING" {
      stats := g[k].(map[string]interface{})["stats"]
      usage := stats.(map[string]interface{})["usage"]
      cpuTotal += usage.(map[string]interface{})["cpu"].(float64)
    }
  }
  cpuAvg = cpuTotal / float64(instCount)

  // log.Println(appName, appURL, instCount, cpuTotal, cpuAvg)

  return instCount, cpuTotal, cpuAvg, nil
}

// func main () {
//
//   instCount, cpuTotal, cpuAvg, err := getAppStats(appName)
//   if err != nil {
//     log.Fatal(err)
//   }
//   log.Println(appName, instCount, cpuTotal, cpuAvg)
// }
