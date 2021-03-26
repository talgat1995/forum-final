package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/astgot/forum/internal/model"
)

// RateHandler ... /rate?post_id= || /rate?comment_id=
func (m *Multiplexer) RateHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rate" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}
		var (
			Islike = true
			isPost = true
		)
		cookie, err := r.Cookie("authenticated")
		var ok bool = true
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
			w.WriteHeader(http.StatusUnauthorized)
			WarnMessage(w, "You need to authorize")
			return
		}
		id, err := strconv.Atoi(r.URL.Query().Get("post_id"))
		if err != nil {
			id, err = strconv.Atoi(r.URL.Query().Get("comment_id"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				WarnMessage(w, "You are walking the wrong way")
				return
			}
			isPost = false

		}
		// if "id" is negative number, it will be dislike
		if id < 0 {
			Islike = false
			id *= -1
		}
		user, err := m.db.GetUserByCookie(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("GetUserByCookie rateHandler.go error")
			WarnMessage(w, "Something went wrong")
			return
		}

		if isPost {
			like := model.NewPostRating()
			like.Like = Islike
			_, err := m.db.GetPostByPID(int64(id))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				WarnMessage(w, "The post not found")
				fmt.Println("GetPostByPID error")
				return
			}
			like.PostID = int64(id)
			// like.UID = user.ID // assign like to UserID --> to check user liked this post,
			// prevent multiple liking of the post
			// Need to return new rate count
			/* Check user liked this post
			if Yes, delete rate from the post
			*/
			isRated := m.db.IsUserRatePost(user.ID, int64(id))
			if isRated {
				if ok := m.db.UpdateRateOfPost(like, user.ID); !ok {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Println("DeleteRateOfPost error")
					WarnMessage(w, "Something went wrong")
					return
				}
			} else {
				if ok := m.db.AddRateToPost(like, user.ID); !ok {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					fmt.Println("AddRatePost() error")
					return
				}
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			like := model.NewCommentRating()
			like.Like = Islike
			comment, err := m.db.GetCommentByID(int64(id))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				WarnMessage(w, "The comment not found")
				fmt.Println("GetCommentByID error")
				return
			}
			like.CommentID = int64(id)
			like.PostID = comment.PostID
			isRated := m.db.IsUserRateComment(user.ID, like.PostID, like.CommentID)
			if isRated {
				if ok := m.db.UpdateRateOfComment(like, user.ID); !ok {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Println("UpdateRateOfComment error")
					WarnMessage(w, "Something went wrong")
					return
				}

			} else {
				if ok := m.db.AddRateToComment(like, user.ID); !ok {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					fmt.Println("AddRateToComment error")
					return
				}
			}
			postID := strconv.Itoa(int(like.PostID))
			http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)

		}

	}
}
