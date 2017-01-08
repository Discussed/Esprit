package Esprit

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	tokenEndpoint         = "https://www.googleapis.com/oauth2/v3/tokeninfo"
	tokenJSONAccess       = "access_token"
	tokenAudience         = ""
	tokenJSONErrorInvalid = "invalid_token"
)

type TokenContainer struct {
	store map[string]*time.Timer
	lock  sync.RWMutex
}

var (
	tokenStore = TokenContainer{}
)

func (container *TokenContainer) Set(token string, ttl int) {
	container.lock.Lock()
	defer container.lock.Unlock()
	if container.store[token] {
		if container.store[token].Stop() {
			container.store[token].Reset(time.Duration(ttl) * time.Second)
		} else {
			<-container.store[token].C
		}
		return
	}
	container.store[token] = time.AfterFunc(
		time.Duration(ttl)*time.Second,
		func() {
			container.lock.Lock()
			defer container.lock.Unlock()
			delete(container.store, token)
		})
}

func (container *TokenContainer) Has(token string) bool {
	container.lock.RLock()
	defer container.lock.RUnlock()
	if container.store[token] {
		return true
	} else {
		return false
	}
}
func (container *TokenContainer) Validate(token string) bool {
	if container.Has(token) {
		return true
	}
	resp, respErr := http.Get(tokenEndpoint + "?" + tokenJSONAccess + "=" + token)
	if respErr != nil {
		log.Fatalln(respErr)
		return false
	}
	defer resp.Body.Close()
	validateResponse := struct {
		Audience string `json:"audience"`
		User     string `json:"user_id"`
		Scope    string `json:"scope"`
		TTL      int    `json:"expires_in"`
		Error    string `json:"error"`
	}{}
	decoder := json.NewDecoder(resp.Body)
	if decodeErr := decoder.Decode(&validateResponse); decodeErr != nil {
		log.Fatalln(decodeErr)
		return false
	}
	if validateResponse.Error == "" {
		if validateResponse.Audience != tokenAudience {
			log.Fatalln("token validation: received audience \"" + validateResponse.Audience + "\" did not match the preset audience \"" + tokenAudience + "\"")
			return false
		}
		container.Set(token, validateResponse.TTL)
	} else if validateResponse.Error == tokenJSONErrorInvalid {
		return false
	} else {
		log.Fatalln("token validation: received error \"" + validateResponse.Error + "\"")
		return false
	}
}
