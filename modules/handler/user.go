package handler

import (
	"coreum_processor/modules/user"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

type KratosIdentity struct {
	Identity string `json:"identity"`
	Email    string `json:"email"`
}

func CreateUser(user *user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		p := KratosIdentity{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"` + `can't parse traits, error:` + err.Error() + `"}`))
			return
		}
		var body []byte
		r.Body.Read(body)
		fmt.Println(string(body))
		err = user.AddUser(p.Identity, "", "")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `can't add user, error:` + err.Error() + `"}`))
			return
		}
	}
}

func LoginUser(userService *user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		p := KratosIdentity{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"` + `can't parse traits, error:` + err.Error() + `"}`))
			return
		}
		_, err = userService.GetUser(p.Identity)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `can't add user, error:` + err.Error() + `"}`))
			return
		}
	}
}
