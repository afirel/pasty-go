package main

import (
  "fmt"
  "net/http"
  "github.com/gorilla/mux"
  "io/ioutil"
  "bytes"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hi there!")
}

func NewSnippetHandler(w http.ResponseWriter, r *http.Request) {
  svc := s3.New(session.New())

  payload, err := ioutil.ReadAll(r.Body)

  if err != nil {
    // Print the error, cast err to awserr.Error to get the Code and
    // Message from an error.
    fmt.Println(err.Error())
    return
  }

  params := &s3.PutObjectInput {
    Bucket: aws.String("pasty-go"),
    Key: aws.String("Test"),
    Body: bytes.NewReader([]byte(payload)),
  }

  resp, err := svc.PutObject(params)

  if err != nil {
    // Print the error, cast err to awserr.Error to get the Code and
    // Message from an error.
    fmt.Println(err.Error())
    return
  }

  fmt.Println(resp)
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler).Methods("GET")
  r.HandleFunc("/snippet", NewSnippetHandler).Methods("POST")
  //r.HandleFunc("/snippet", SnippetHandler).Methods("GET")

  http.Handle("/", r)
  http.ListenAndServe(":8080", nil)
}
