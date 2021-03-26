package controller

import (
	"net/http"
	"strconv"

	"github.com/astgot/forum/internal/model"
)

// FilterHandler ...
/* To demonstrate:
   1) created posts --> /filter?section=my_posts
   2) liked posts   --> /filter?section=rated
   3) by categories --> /filter?section=by_category
*/
func (m *Multiplexer) FilterHandler() http.HandlerFunc {
	type PostRaw struct {
		Post       *model.Post
		Categories []string
		PostRate   *model.PostRating
	}
	var filter struct {
		Section    string
		AuthUser   *model.Users
		PostScroll []*PostRaw
		Error      string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/filter" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}

		section := r.URL.Query().Get("section")
		if section == "my_posts" {
			c, err := r.Cookie("authenticated")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				WarnMessage(w, "You need to authorize")
				return
			} else if err == nil {
				ok, err := m.db.IsCookieInDB(c.Value)
				if err != nil {
					WarnMessage(w, "Something went wrong")
					return
				}
				// if db doesn't have cookie --> redirect to main page
				if !ok {
					delCookie(w)
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}
			user, err := m.db.GetUserByCookie(c.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			// Add function to find posts of users by their ID
			/* 1) retrive posts from table "Posts"
			 */
			posts, err := m.db.GetPostsByUID(user.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			for _, post := range posts {
				result := &PostRaw{}
				result.Post = post
				result.Categories, err = m.db.GetCategoryOfPost(post.ID)
				result.PostRate = m.db.GetRateCountOfPost(post.ID)
				filter.PostScroll = append(filter.PostScroll, result)
			}
			filter.Section = "My Posts"
			filter.AuthUser = user
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "filter.html", filter)
			//Prevent from doubling of content
			filter.Section = ""
			filter.AuthUser = nil
			filter.PostScroll = nil
			return
		} else if section == "liked" {
			c, err := r.Cookie("authenticated")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				WarnMessage(w, "You need to be authorized")
				return
			} else if err == nil {
				ok, err := m.db.IsCookieInDB(c.Value)
				if err != nil {
					WarnMessage(w, "Something went wrong")
					return
				}
				// if db doesn't have cookie --> redirect to main page
				if !ok {
					delCookie(w)
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}
			user, err := m.db.GetUserByCookie(c.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			// Add function to find post rated by the user
			/* 1) retrieve liked posts from "RateUserPost"
			 */
			posts, err := m.db.GetRatedPostsByUID(user.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			for _, post := range posts {
				result := &PostRaw{}
				result.Post = post
				result.Categories, err = m.db.GetCategoryOfPost(post.ID)
				result.PostRate = m.db.GetRateCountOfPost(post.ID)
				filter.PostScroll = append(filter.PostScroll, result)

			}
			filter.Section = "Liked posts"
			filter.AuthUser = user
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "filter.html", filter)
			filter.Section = ""
			filter.AuthUser = nil
			filter.PostScroll = nil
			return
		} else if section == "by_category" {
			// Filter by category
			if r.Method == "POST" {
				r.ParseForm()
				search := r.FormValue("categories")
				c, err := r.Cookie("authenticated")
				if err == nil {
					ok, err := m.db.IsCookieInDB(c.Value)
					if err != nil {
						WarnMessage(w, "Something went wrong")
						return
					}
					if !ok {
						delCookie(w)
					} else {
						user, _ := m.db.GetUserByCookie(c.Value)
						filter.AuthUser = user
					}
				}
				/*
				   1)search posts in the table "PostMapping" (retrieve all postID)
				   2)show posts from table "Posts"
				*/
				posts, err := m.db.GetPostsByCategories(search)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				// If search doesn't give any results
				if len(posts) == 0 {
					filter.Error = "No results"
				} else {
					filter.Error = "Number of posts: " + strconv.Itoa(len(posts))

				}
				for _, post := range posts {
					result := &PostRaw{}
					result.Post = post
					result.Categories, _ = m.db.GetCategoryOfPost(post.ID)
					result.PostRate = m.db.GetRateCountOfPost(post.ID)
					filter.PostScroll = append(filter.PostScroll, result)
				}
				w.WriteHeader(http.StatusOK)
				tpl.ExecuteTemplate(w, "filter.html", filter)
			}
			filter.AuthUser = nil
			filter.Section = ""
			filter.PostScroll = nil
			filter.Error = ""
		} else {
			w.WriteHeader(http.StatusBadRequest)
			WarnMessage(w, "Bad Request")
			return
		}
	}
}
