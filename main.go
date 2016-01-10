package main

import (
  "os"
  "fmt"
  "net/http"
  "crypto/md5"
  "encoding/hex"
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
    handleError(err, w)
    return
  }

  hasher := md5.New()
  hasher.Write([]byte(payload))
  md5sum := hex.EncodeToString(hasher.Sum(nil))

  params := &s3.PutObjectInput {
    Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
    Key: aws.String(md5sum),
    Body: bytes.NewReader([]byte(payload)),
  }

  resp, err := svc.PutObject(params)

  if err != nil {
    handleError(err, w)
    return
  }

  fmt.Println(resp)
  fmt.Println(md5sum)

  // construct URL to new object
  var prefix = os.Getenv("URL_PREFIX")
  if prefix == "" {
    prefix = fmt.Sprintf("http://%s", r.Host)
  }
  url := fmt.Sprintf("%s/snippet/%s", prefix, md5sum)
  fmt.Fprintf(w, "%s", url)
}

func handleError(err error, w http.ResponseWriter) {
  fmt.Println(err.Error())
  http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler).Methods("GET")
  r.HandleFunc("/snippet", NewSnippetHandler).Methods("POST")
  //r.HandleFunc("/snippet", SnippetHandler).Methods("GET")

  http.Handle("/", r)
  http.ListenAndServe(":8080", nil)
}
