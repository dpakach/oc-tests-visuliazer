package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
)
type scenarios []string
type issueData map[string]scenarios
const urlRegex = "https:\\/\\/[a-zA-Z0-9]+\\.[^\\s]{2,}(\\/[a-zA-Z0-9\\-]+)+"
const scenarioRegex = "[a-zA-Z0-9]+\\/[a-zA-Z0-9]+\\.feature:[0-9]+"

const day = time.Second * 60 * 60 * 24

var urlR = regexp.MustCompile(urlRegex)
var scenarioR = regexp.MustCompile(scenarioRegex)

var storages = []string{"OWNCLOUD", "OCIS"}

func (d *Data)updateData() {
  for {
    fmt.Println("updating data")
    for _, storage := range storages {
      storageData := issueData{}
      storageDataSuite := issueData{}

      reader, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/owncloud/ocis/master/ocis/tests/acceptance/expected-failures-on-%v-storage.txt", storage))
      if err != nil {
        fmt.Println(err)
        continue
      }
      defer reader.Body.Close()

      scanner := bufio.NewScanner(reader.Body)

      currentUrl := ""
      scenariosCount := 0
      for scanner.Scan() {
        matches := urlR.FindAllString(string(scanner.Text()), -1)

        if len(matches) > 0 {
          currentUrl = matches[0]
          if scenariosCount < 1 {
            continue
          }
          _, ok := storageData[matches[0]]
          if !ok {
            storageData[matches[0]] = []string{}
            currentUrl = matches[0]
            scenariosCount = 0
          }
        } else {
          matches := scenarioR.FindAllString(string(scanner.Text()), -1)
          parts := strings.Split(scanner.Text(), "/")

          if scanner.Text() != "" && scanner.Text()[0] != '#' && len(parts) > 0 {
            suite := parts[0]
            suiteData, ok := storageDataSuite[suite]
            if !ok {
              storageDataSuite[suite] = []string{scanner.Text()}
            } else {
              storageDataSuite[suite] = append(suiteData, scanner.Text())
            }
          }

          if currentUrl == "" {
            continue
          }
          if len(matches) > 0 {
            storageData[currentUrl] = append(storageData[currentUrl], scanner.Text())
            scenariosCount += 1
          }
        }
      }
      if storage == "OCIS" {
        d.ocisData = &storageData
        d.ocisSuiteData = &storageDataSuite
      } else {
        d.ocData = &storageData
        d.ocSuiteData = &storageDataSuite
      }
    }
    time.Sleep(day)
  }
}

func main() {
  l := log.New(os.Stdout, "oc-issue-struct", log.LstdFlags)

	dh := NewData(&issueData{}, &issueData{}, &issueData{}, &issueData{})

  go dh.updateData()

	// create a new serve mux and register the handlers
	sm := http.NewServeMux()
	sm.Handle("/api", dh)

  sm.Handle("/", http.FileServer(http.Dir("./public")))

	// create a new server
	s := http.Server{
		Addr:         ":8880",           // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     l,                 // set the logger for the server
		ReadTimeout:  5 * time.Second,   // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}

	// start the server
	go func() {
		l.Println("Starting server on port 8880")

		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}

type Data struct {
  ocData *issueData
  ocisData *issueData
  ocSuiteData *issueData
  ocisSuiteData *issueData
}

func NewData(ocData, ocisData, ocSuiteData, ocisSuiteData *issueData) *Data {
	return &Data{ocData, ocisData, ocSuiteData, ocisSuiteData}
}

func (d *Data) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
  keys := r.URL.Query()
  var storage string
  var by string

  val, ok := keys["by"]
  if !ok || len(val[0]) < 1 {
    by = "issue"
  } else {
    by = val[0]
  }

  val, ok = keys["storage"]

  if !ok || len(val[0]) < 1 {
    storage = "oc"
  } else {
    storage = val[0]
  }

  if by != "" {
    switch by {
    case "suite":
    switch storage {
      case "ocis":
        res, err := json.Marshal(d.ocisSuiteData)
        if err != nil {
          rw.WriteHeader(500)
        }
        rw.Write(res)
      default:
        res, err := json.Marshal(d.ocSuiteData)
        if err != nil {
          rw.WriteHeader(500)
        }
        rw.Write(res)
      }
    default:
      switch storage {
      case "ocis":
        res, err := json.Marshal(d.ocisData)
        if err != nil {
          rw.WriteHeader(500)
        }
        rw.Write(res)
      default:
        res, err := json.Marshal(d.ocData)
        if err != nil {
          rw.WriteHeader(500)
        }
        rw.Write(res)
      }
    }
  } else {
    switch storage {
    case "ocis":
      res, err := json.Marshal(d.ocisData)
      if err != nil {
        rw.WriteHeader(500)
      }
      rw.Write(res)
    default:
      res, err := json.Marshal(d.ocData)
      if err != nil {
        rw.WriteHeader(500)
      }
      rw.Write(res)
    }
  }
}
