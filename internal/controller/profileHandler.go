package controller

import (
	"net/http"
)

// ProfileHandler ..
func (m *Multiplexer) ProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("authenticated")
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			WarnMessage(w, "You need to login")
			return
		}
		user, err := m.db.GetUserByCookie(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			WarnMessage(w, "Cookie is expired")
			return
		}
		w.WriteHeader(http.StatusOK)
		tpl.ExecuteTemplate(w, "profile.html", user)
	}
}
