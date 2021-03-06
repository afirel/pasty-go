package main

import (
  "os"
  "fmt"
  "net/http"
  "crypto/md5"
  "encoding/hex"
  "github.com/gorilla/mux"
  "io"
  "io/ioutil"
  "bytes"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/awserr"
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
    handleError(err, w, http.StatusInternalServerError)
    return
  }

  objectId := md5sum(payload)

  params := &s3.PutObjectInput {
    Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
    Key: aws.String(objectId),
    Body: bytes.NewReader([]byte(payload)),
  }

  resp, err := svc.PutObject(params)

  if err != nil {
    handleError(err, w, http.StatusInternalServerError)
    return
  }

  fmt.Println(resp)
  fmt.Println(objectId)

  fmt.Fprintf(w, "%s", urlFor(objectId, r))
}

func md5sum(payload []byte) string {
  hasher := md5.New()
  hasher.Write(payload)
  return hex.EncodeToString(hasher.Sum(nil))
}

func urlFor(objectId string, r *http.Request) string {
  var prefix = os.Getenv("URL_PREFIX")
  if prefix == "" {
    prefix = fmt.Sprintf("http://%s", r.Host)
  }
  return fmt.Sprintf("%s/%s", prefix, objectId)
}

func SnippetHandler(w http.ResponseWriter, r *http.Request) {
  id := mux.Vars(r)["id"]

  svc := s3.New(session.New())
  params := &s3.GetObjectInput {
    Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
    Key: aws.String(id),
  }
  resp, err := svc.GetObject(params)
  if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
      if awsErr.Code() == "NoSuchKey" {
        handleError(err, w, http.StatusNotFound)
      } else {
        handleError(err, w, http.StatusInternalServerError)
      }
    } else {
      handleError(err, w, http.StatusInternalServerError)
    }
    return
  }
  fmt.Println(resp)
  io.Copy(w, resp.Body)
}

func handleError(err error, w http.ResponseWriter, code int) {
  fmt.Println(err.Error())
  http.Error(w, err.Error(), code)
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler).Methods("GET")
  r.HandleFunc("/", NewSnippetHandler).Methods("POST")
  r.HandleFunc("/{id}", SnippetHandler).Methods("GET")

  http.Handle("/", r)
  http.ListenAndServe(":80", nil)
}
