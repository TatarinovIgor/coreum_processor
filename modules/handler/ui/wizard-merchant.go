package ui

import (
	"context"
	"coreum_processor/modules/internal"
	"coreum_processor/modules/user"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"log"
	"net/http"
)

func PageWizardMerchant(ctx context.Context, userService *user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}
		userStore = userStore
		t, err := template.ParseFiles("./templates/lite/default/onboarding-wizard.html",
			"./templates/lite/default/ory-kratos-form.html")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
		query := r.URL.Query()
		step := query.Get("step")
		if step == "" {
			step = "001"
		}
		err = t.ExecuteTemplate(w, "onboarding-wizard.html", wizardPages[step])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"` + `template parsing error` + `"}`))
			return
		}
	}
}

func PageWizardMerchantUpdate(ctx context.Context, userService *user.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		userStore, err := internal.GetUserStore(r.Context())
		if err != nil {
			log.Println(`can't find user store`)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"` + `can't find user store` + `"}`))
			return
		}
		query := r.URL.Query()
		step := query.Get("step")
		err = r.ParseForm()
		if err != nil {
			log.Println(`can't parse form, error: `, err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"` + `parser form, error: ` + err.Error() + `"}`))
			return
		}
		switch step {
		case "001":
			userStore.FirstName = r.Form.Get("first_name")
			userStore.LastName = r.Form.Get("last_name")
			err := userService.UpdateUser(*userStore)
			if err != nil {
				http.Redirect(w, r, r.URL.Path+"?step=001", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, r.URL.Path+"?step=002", http.StatusSeeOther)
			}
		case "002":
			merchID := r.Form.Get("merchant_id")
			err := userService.LinkUserToMerchant(userStore.Identity, merchID)
			if err != nil {
				http.Redirect(w, r, r.URL.Path+"?step=002", http.StatusSeeOther)
			} else {
				_, err = userService.SetUserAccess(userStore.Identity, user.SetOnboarded(userStore.Access))
				if err != nil {
					log.Println(err)
				}
				http.Redirect(w, r, "/ui/dashboard", http.StatusSeeOther)
			}
		case "003":
			merchantName := r.Form.Get("merchant_name")
			merchantEmail := r.Form.Get("merchant_email")
			err = userService.RequestMerchantForUser(userStore.Identity, merchantName, merchantEmail)
			if err != nil {
				log.Println(err)
				http.Redirect(w, r, r.URL.Path+"?step=002", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, r.URL.Path+"?step=004", http.StatusSeeOther)
			}
		default:
			log.Println("unknown wizard step: ", step)
		}
	}
}

var (
	isTrue      = true
	isFalse     = false
	wizardPages = WizardPages{
		"001": WizardPage{
			Method: "POST",
			Action: "/ui/merchant/onboarding-wizard?step=001",
			Messages: []WizardText{
				{Text: "Fill your personal data"},
			},
			Nodes: []WizardNode{
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "First Name",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "first_name",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "input",
							Value:        nil,
						},
					},
				},
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Last Name",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "last_name",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "input",
							Value:        nil,
						},
					},
				},
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Submit",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "Submit",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "submit",
							Value:        nil,
						},
					},
				},
			},
		},
		"002": WizardPage{
			Method: "POST",
			Action: "/ui/merchant/onboarding-wizard?step=002",
			Messages: []WizardText{
				{Text: "Chose Merchant to work with"},
			},
			Nodes: []WizardNode{
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Merchant ID",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "merchant_id",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "input",
							Value:        nil,
						},
					},
				},
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Submit",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "Submit",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "submit",
							Value:        nil,
						},
					},
				},
				{
					Type: "a",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Merchant ID",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeAnchorAttributes: &WizardNodeAnchorAttributes{
							Href:  "/ui/merchant/onboarding-wizard?step=003",
							Title: WizardText{Text: "I would like to make new merchant"},
						},
					},
				},
			},
		},
		"003": WizardPage{
			Method: "POST",
			Action: "/ui/merchant/onboarding-wizard?step=003",
			Messages: []WizardText{
				{Text: "Provide Merchant details for registration"},
			},
			Nodes: []WizardNode{
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Merchant Name",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "merchant_name",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "input",
							Value:        nil,
						},
					},
				},
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Merchant Email",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "merchant_email",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "input",
							Value:        nil,
						},
					},
				},
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Submit",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "Submit",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "submit",
							Value:        nil,
						},
					},
				},
			},
		},
		"004": WizardPage{
			Method: "GET",
			Action: "/",
			Messages: []WizardText{
				{Text: "Thank you for provided information"},
				{Text: "Your request has accepted and pending approval"},
			},
			Nodes: []WizardNode{
				{
					Type: "input",
					Meta: WizardNodeMeta{Label: &WizardText{
						Id:   0,
						Text: "Ok",
						Type: "",
					}},
					Attributes: WizardNodeAttributes{
						UiNodeInputAttributes: &WizardNodeInputAttributes{
							Autocomplete: nil,
							Disabled:     false,
							Label:        nil,
							Name:         "Ok",
							NodeType:     "input",
							Onclick:      nil,
							Pattern:      nil,
							Required:     &isTrue,
							Type:         "submit",
							Value:        nil,
						},
					},
				},
			},
		},
	}
)
