package controller

import (
	"fmt"
	"net/http"

	"github.com/astgot/forum/internal/model"
)

// MainHandle ...
func (m *Multiplexer) MainHandle() http.HandlerFunc {

	// Need to create structure to show array of Users, Posts, Comments, Categories for arranging them in HTML
	type PostRaw struct {
		Post       *model.Post
		Categories []string
		PostRate   *model.PostRating
	}
	var mainPage struct {
		AuthUser   *model.Users
		PostScroll []*PostRaw
	}
	// Here we can create our own struct, which is usable only here
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/main" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}
		posts := m.GetAllPosts(w)

		cookie, err := r.Cookie("authenticated")
		var ok bool = true // for cookie
		if err == nil {
			ok, err1 := m.db.IsCookieInDB(cookie.Value)
			if err1 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			if !ok {
				delCookie(w)
			}
		}
		if !ok || err != nil {
			// if user is guest, retrieve all posts for displaying
			for _, post := range posts {
				guest := &PostRaw{}
				// guest.Comments, _ = m.db.GetCommentsOfPost(post.PostID)
				// guest.Author, _ = m.db.FindByUserID(post.UserID)
				guest.Post = post
				guest.Categories, err = m.db.GetCategoryOfPost(post.ID)
				guest.PostRate = m.db.GetRateCountOfPost(post.ID)
				mainPage.PostScroll = append(mainPage.PostScroll, guest)
			}
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "main.html", mainPage)
			mainPage.PostScroll = nil
			return
		}
		// if User is authenticated
		user, err := m.db.GetUserByCookie(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			WarnMessage(w, "Something went wrong")
			return
		}
		mainPage.AuthUser = user
		for _, post := range posts {
			auth := &PostRaw{}
			auth.Post = post
			auth.Categories, err = m.db.GetCategoryOfPost(post.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				fmt.Println("Threads retrieving error")
				return
			}
			auth.PostRate = m.db.GetRateCountOfPost(post.ID)
			mainPage.PostScroll = append(mainPage.PostScroll, auth)
		}
		w.WriteHeader(http.StatusOK)
		tpl.ExecuteTemplate(w, "main.html", mainPage)
		// prevent from posts doubling
		mainPage.AuthUser = nil
		mainPage.PostScroll = nil
	}
}
