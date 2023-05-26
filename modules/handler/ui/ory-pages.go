package ui

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/client-go"
	"html/template"
	"log"
	"net/http"
)

func GetPageSignUp(ctx context.Context, ory *client.APIClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var execute *client.RegistrationFlow
		var resp *http.Response

		q := r.URL.Query()
		if flow, ok := q["flow"]; !ok {
			execute, resp, err := ory.FrontendApi.CreateBrowserRegistrationFlow(ctx).Execute()
			if err != nil {
				log.Println(err, resp)
				return
			}
			http.Redirect(w, r, execute.RequestUrl, http.StatusSeeOther)
			return
		} else {
			// set the cookies on the ory client
			cookies := r.Header.Get("Cookie")
			// check if we have a session
			session, _, err := ory.FrontendApi.ToSession(ctx).Cookie(cookies).Execute()
			if (err == nil && session != nil) || (err == nil && *session.Active) {
				// this will redirect the user to the managed Ory Login UI
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				//http.Redirect(writer, request, "/.ory/self-service/login/browser", http.StatusSeeOther)
				return
			}
			execute, resp, err = ory.FrontendApi.GetRegistrationFlow(ctx).Cookie(cookies).Id(flow[0]).Execute()
			if err != nil {
				log.Println(err, resp)
				return
			}
		}

		t, err := template.ParseFiles("./templates/lite/default/auth-signup.html",
			"./templates/lite/default/ory-kratos-form.html")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		err = t.ExecuteTemplate(w, "auth-signup.html", execute.Ui)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func GetPageLogIn(ctx context.Context, ory *client.APIClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var execute *client.LoginFlow
		var err error

		var resp *http.Response
		q := r.URL.Query()
		if flow, ok := q["flow"]; !ok {
			execute, resp, err = ory.FrontendApi.CreateBrowserLoginFlow(ctx).Execute()
			if err != nil {
				log.Println("can't make flow for login, error: ", err)
				return
			}
			http.Redirect(w, r, execute.RequestUrl, http.StatusSeeOther)
			return
		} else {
			// set the cookies on the ory client
			var cookies string

			// this example passes all request.Cookies
			// to `ToSession` function
			//
			// However, you can pass only the value of
			// ory_session_projectid cookie to the endpoint
			cookies = r.Header.Get("Cookie")
			execute, resp, err = ory.FrontendApi.GetLoginFlow(ctx).Cookie(cookies).Id(flow[0]).Execute()
		}
		if err != nil {
			log.Println(err, resp)
			return
		}

		t, err := template.ParseFiles("./templates/lite/default/auth-signin.html",
			"./templates/lite/default/ory-kratos-form.html")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		err = t.ExecuteTemplate(w, "auth-signin.html", execute.Ui)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}
func GetPageLogOut(ctx context.Context, ory *client.APIClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// this example passes all request.Cookies
		// to `ToSession` function
		//
		// However, you can pass only the value of
		// ory_session_projectid cookie to the endpoint
		var cookies string
		cookies = r.Header.Get("Cookie")
		execute, _, err := ory.FrontendApi.CreateBrowserLogoutFlow(ctx).Cookie(cookies).Execute()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		if execute != nil {
			http.Redirect(w, r, execute.LogoutUrl, http.StatusSeeOther)
		}
	}
}

func PageReset(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	t, err := template.ParseFiles("./templates/lite/default/password-reset.html") //ToDo finish page
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}
}

func PageError(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	t, err := template.ParseFiles("./templates/lite/default/password-reset.html") //ToDo finish page
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
		return
	}
}
