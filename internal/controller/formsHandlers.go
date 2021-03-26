package controller

import (
	"fmt"
	"net/http"

	"github.com/astgot/forum/internal/model"
)

// SignupHandle ---> /signup
func (m *Multiplexer) SignupHandle() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/signup" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}
		c, err := r.Cookie("authenticated")
		if err == nil {
			ok, err := m.db.IsCookieInDB(c.Value)
			if err != nil {
				WarnMessage(w, "Something went wrong")
				return
			}
			// if db has cookie redirect to main page
			if ok {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				// if not, delete it from client side
			} else {
				delCookie(w)
			}
			return
		}
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "signup.html", nil)
		} else if r.Method == "POST" {
			r.ParseForm() // Parsing Form from the front-end
			user := &model.Users{
				Firstname:  r.PostFormValue("Firstname"),
				Lastname:   r.PostFormValue("Lastname"),
				Username:   r.PostFormValue("Username"),
				Email:      r.PostFormValue("Email"),
				Password:   r.PostFormValue("Password"),
				ConfirmPwd: r.PostFormValue("Confirm"),
			}

			if ValidateInput(user) == false {
				tpl.ExecuteTemplate(w, "signup.html", user)
				return
			}

			encryptPass := HashPassword(user.Password)
			user.EncryptedPwd = encryptPass   // fill with Encrypted Password
			newUser, err := m.db.Create(user) // Sending
			if err != nil {
				// Output errors, if username or e-mail is busy
				user.Errors = make(map[string]string)
				if err.Error() == "UNIQUE constraint failed: Users.email" {
					user.Errors["Email"] = "That e-mail is already taken, please use another"
				} else if err.Error() == "UNIQUE constraint failed: Users.username" {
					user.Errors["Username"] = "That username is already taken, please use another"
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				tpl.ExecuteTemplate(w, "signup.html", user)
				return
			}
			m.AddSession(w, "guest", newUser) // guest session
			//w.WriteHeader(http.StatusOK)
			http.Redirect(w, r, "/confirmation", http.StatusSeeOther)

		}
	}

}

// LoginHandle ---> /login
func (m *Multiplexer) LoginHandle() http.HandlerFunc {
	type Login struct {
		auth         bool
		unameOrEmail bool
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}
		c, err := r.Cookie("authenticated")
		if err == nil {
			ok, err := m.db.IsCookieInDB(c.Value)
			if err != nil {
				WarnMessage(w, "Something went wrong")
				return
			}
			// if db has cookie redirect to main page
			if ok {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				// if not, delete it from client side
			} else {
				delCookie(w)
			}
			return
		}

		if r.Method == "GET" {
			tpl.ExecuteTemplate(w, "login.html", nil)
		} else if r.Method == "POST" {
			r.ParseForm()
			login := &model.Users{
				Username: r.PostFormValue("Username"),
				Password: r.PostFormValue("Password"),
			}
			check := Login{}

			check.unameOrEmail = UnameOrEmail(login.Username)

			if check.unameOrEmail {
				u, err := m.db.FindByEmail(login.Username)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					WarnMessage(w, "Invalid credentials")
					return
				}
				check.auth = ComparePassword(u.EncryptedPwd, login.Password)
				if !check.auth {
					w.WriteHeader(http.StatusBadRequest)
					WarnMessage(w, "Invalid credentials")
					return
				}
			} else if !check.unameOrEmail {
				u, err := m.db.FindByUsername(login.Username)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					WarnMessage(w, "Invalid credentials")
					return
				}
				check.auth = ComparePassword(u.EncryptedPwd, login.Password)
				if !check.auth {
					w.WriteHeader(http.StatusBadRequest)
					WarnMessage(w, "Invalid credentials")
					return
				}

			}
			login.ID = m.db.GetUserID(login, check.unameOrEmail)
			err = m.db.IsUserAuthenticated(login)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("Error on deleting cookie:")
				WarnMessage(w, "Something went wrong")
				return
			}
			m.AddSession(w, "authenticated", login) // Add cookie session after successful authentication
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

	}
}

// ConfirmHandler --> /confirmation
func (m *Multiplexer) ConfirmHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/confirmation" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}
		c, err := r.Cookie("authenticated")
		if err == nil {
			ok, err := m.db.IsCookieInDB(c.Value)
			if err != nil {
				WarnMessage(w, "Something went wrong")
				return
			}
			if ok {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			} else {
				delCookie(w)
			}
			return
		}
		tpl.ExecuteTemplate(w, "confirmation.html", nil)
	}
}

// LogoutHandle ...
func (m *Multiplexer) LogoutHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logout" {
			cookie, err := r.Cookie("authenticated")
			if err != nil {
				m.AddSession(w, "guest", nil)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
				/*OR http.Error()*/
			}
			m.DeleteSession(w, cookie.Value)
			//w.WriteHeader(http.StatusOK)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

	}
}
