package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type UrlStore struct {
	items        map[string]string
	index        uint64
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

func (store *UrlStore) Add(url string) string {
	store.Lock()
	key := fmt.Sprintf("u%d", store.index)
	store.items[key] = url
	store.index += 1
	store.Unlock()
	return key
}

func (store *UrlStore) Get(key string) (string, bool) {
	store.RLock()
	val, ok := store.items[key]
	store.RUnlock()
	return val, ok
}

func NewUrlStore() *UrlStore {
	store := UrlStore{items: make(map[string]string), index: 0}
	return &store
}

type StoreUrlRequest struct {
	Url string `json:"url"`
}

type StoreUrlResponse struct {
	Key string `json:"key"`
}

var urlStore = NewUrlStore()

func storeUrlHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var storeRequest StoreUrlRequest

	err := decoder.Decode(&storeRequest)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	key := urlStore.Add(storeRequest.Url)

	storeResponse := StoreUrlResponse{Key: key}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(storeResponse)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func retrieveUrlHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	key := params["key"]
	url, ok := urlStore.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", storeUrlHandler).Methods("POST")
	router.HandleFunc("/{key}", retrieveUrlHandler).Methods("GET")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
