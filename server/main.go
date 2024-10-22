package main

import (
	"crypto"
	"encoding/pem"
	"fmt"
	"goenv/api"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"crypto/x509"

	"github.com/fullsailor/pkcs7"
	"github.com/joho/godotenv"
)

var env, _ = godotenv.Read(".env")

type RenderData struct {
	ProfileIDs string `json:"profile-ids"` // comma seperated
}

func main() {
	fmt.Println(":: Enter program ::")

	err := api.Connect()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to MySQL database")
	defer api.Close()

	// !IMPORTANT expected query codes=["STARB", "OTHER"]

	http.HandleFunc("GET /", Root)
	http.HandleFunc("GET /CA", CA)
	http.HandleFunc("GET /enroll", GetEnroll)
	http.HandleFunc("POST /enroll", PostEnroll)

	// TODO: replace self signed certs with something from: https://support.apple.com/en-us/103272 (list of supported CAs that don't ask the user if they trust it)
	err = http.ListenAndServeTLS(":8080", env["CERT"], env["KEY"], nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

func Render(w http.ResponseWriter, r *http.Request, templateFilePath string, data any) bool {
	t, err := template.ParseFiles(templateFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println("Server template error:", err)
		w.Write([]byte(err.Error()))
		return false
	}

	return true
}

func Root(w http.ResponseWriter, r *http.Request) {
	Render(w, r, "html/root.gohtml", RenderData{
		ProfileIDs: r.URL.Query().Get("profile-ids"),
	})
}

func CA(w http.ResponseWriter, r *http.Request) {
	fileInfo, err := os.Stat(env["CERT"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=" + strconv.Quote(fileInfo.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, env["CERT"])
}

func GetEnroll(w http.ResponseWriter, r *http.Request) {
	if !Render(w, r, "html/login.gohtml", RenderData{ ProfileIDs: r.URL.Query().Get("profile-ids"), }) {
		return
	}
}

func PostEnroll(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Enrollment http request error", http.StatusInternalServerError)
		return
	}

	if !r.Form.Has("username") || !r.Form.Has("password") {
		http.Error(w, "Enrollment http request error", http.StatusInternalServerError)
		return
	}

	// TODO: limit username password to 255 characters
	_, err = api.AuthorizeUser(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/x-apple-aspen-config")

	config, err := api.ProfileServicePayload(r, env["CHALLENGE"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p7, err := pkcs7.NewSignedData(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cert, key, err := loadCertificate(env["CERT"], env["KEY"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = p7.AddSigner(cert, key, pkcs7.SignerInfoConfig{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signedData, err := p7.Finish()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(signedData)
}

func loadCertificate(certFile, keyFile string) (*x509.Certificate, crypto.PrivateKey, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate: %v", err)
	}

	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key: %v", err)
	}

	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Parse private key
	block, _ = pem.Decode(keyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode PEM block containing key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return cert, key, nil
}