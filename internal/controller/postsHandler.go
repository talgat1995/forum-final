package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/astgot/forum/internal/model"
	uuid "github.com/satori/go.uuid"
)

const maxUploadImage = 20 * 1024 * 1024 // 20 MB

// CreatePostHandler ...
func (m *Multiplexer) CreatePostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/create" {
			w.WriteHeader(http.StatusNotFound)
			WarnMessage(w, "404 Not Found")
			return
		}

		// Check user authorization
		cookie, err := r.Cookie("authenticated")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err == nil {
			ok, err := m.db.IsCookieInDB(cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				WarnMessage(w, "Something went wrong")
				return
			}
			if !ok {
				delCookie(w)
			}
		}

		u, err := m.db.GetUserByCookie(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			WarnMessage(w, "Something went wrong")
			fmt.Println("GetUserByCookie error")
			return
		}
		var Create struct {
			Errors   map[string]string
			Username string
		}
		Create.Username = u.Username
		// Gathering post data
		post := model.NewPost()
		category := model.NewCategory()
		if r.Method == "POST" {
			Create.Errors = make(map[string]string)
			// image uploading
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadImage)
			if err := r.ParseMultipartForm(maxUploadImage); err != nil {
				Create.Errors["Image"] = "The uploaded image is too big. Please choose another image"
				tpl.ExecuteTemplate(w, "postCreate.html", Create)
				return
			}
			file, fileHeader, err := r.FormFile("picture")
			// if file has chosen
			if err == nil {
				defer file.Close()
				buff := make([]byte, 512)
				_, err = file.Read(buff)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				fileType := http.DetectContentType(buff)
				if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/jpg" && fileType != "image/gif" {
					Create.Errors["Image"] = "The provided image format is not allowed. Please upload a JPEG, PNG, GIF image"
					w.WriteHeader(http.StatusBadRequest)
					tpl.ExecuteTemplate(w, "postCreate.html", Create)
					return
				}
				_, err = file.Seek(0, io.SeekStart)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				err = os.MkdirAll("./uploads", os.ModePerm) // makes uploads directory if it doesnt exist
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				imageName := uuid.NewV4().String()
				destImage := fmt.Sprintf("./uploads/%s%s", imageName, filepath.Ext(fileHeader.Filename))
				dst, err := os.Create(destImage)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				defer dst.Close()
				_, err = io.Copy(dst, file) // copies the file content into the new one in the uploads folder
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					WarnMessage(w, "Something went wrong")
					return
				}
				post.Image = destImage
			}
			r.ParseForm()
			post.UserID = u.ID
			post.Author = u.Firstname + " " + u.Lastname + " aka " + "\"" + u.Username + "\""
			post.Title = r.PostFormValue("title")
			post.Content = r.PostFormValue("postContent")
			category.Name = r.Form["category"]
			if post.Title == "" {
				Create.Errors["Title"] = "\"Title\" field is empty"
				tpl.ExecuteTemplate(w, "postCreate.html", Create)
				return
			} else if len(post.Title) > 20 {
				Create.Errors["Title"] = "\"Title\" is too long"
				tpl.ExecuteTemplate(w, "postCreate.html", Create)
				return
			} else if post.Content == "" {
				Create.Errors["Content"] = "\"Content\" field is empty"
				tpl.ExecuteTemplate(w, "postCreate.html", Create)
				return
			} else if len(category.Name) == 0 {
				Create.Errors["Category"] = "Please select one or more category"
				tpl.ExecuteTemplate(w, "postCreate.html", Create)
				return
			}

			// get categories from checkbox (may be more than 1)
			//threads := CheckNumberOfThreads(thread.Name)
			post.CreationDate = time.Now().Format("January 2 15:04")
			post.ID, err = m.db.InsertPostInfo(post)
			if post.ID == -1 {
				w.WriteHeader(http.StatusBadRequest)
				WarnMessage(w, "Bad request")
				return
			} else if err != nil {
				fmt.Println(err.Error())
				return
			}
			// If post has several threads, to this post will attach this info
			for _, categoryName := range category.Name {
				if err = m.db.InsertPostMapInfo(post.ID, categoryName); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Println("InsertPostMapInfo error:", err.Error())
					WarnMessage(w, "Something went wrong")
				}
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)

		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "postCreate.html", Create)
		}

	}
}

// PostView ... (/post?id=)
//(single post viewing -> to see comments, rate count OR if user is authenticated, he able to add comments and rate post here)
func (m *Multiplexer) PostView() http.HandlerFunc {

	type PostAttr struct {
		Categories []string
		Comments   []*model.Comments
	}
	var singlePost struct {
		PostInfo []*PostAttr
		AuthUser *model.Users
		Post     *model.Post
		Errors   string
	}
	return func(w http.ResponseWriter, r *http.Request) {

		id, errID := strconv.Atoi(r.URL.Query().Get("id"))
		if errID != nil {
			w.WriteHeader(http.StatusBadRequest)
			WarnMessage(w, "Invalid input id")
			return
		}

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
			// If user is guest
			postAttr := &PostAttr{}
			singlePost.Post, err = m.db.GetPostByPID(int64(id))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				fmt.Println("Error on PostView() function")
				WarnMessage(w, "The post not found")
				return
			}
			postAttr.Comments, err = m.db.GetCommentsOfPost(int64(id))
			commentRate := model.NewCommentRating()
			for i := 0; i < len(postAttr.Comments); i++ {
				commentRate = m.db.GetRateCountOfComment(postAttr.Comments[i].CommentID, int64(id))
				postAttr.Comments[i].LikeCnt = commentRate.LikeCount
				postAttr.Comments[i].DislikeCnt = commentRate.DislikeCount
			}
			postAttr.Categories, err = m.db.GetCategoryOfPost(int64(id))
			singlePost.PostInfo = append(singlePost.PostInfo, postAttr)
			singlePost.Post.ID = int64(id)
			w.WriteHeader(http.StatusOK)
			tpl.ExecuteTemplate(w, "postView.html", singlePost)
			singlePost.Post = nil
			singlePost.PostInfo = nil
			return
		}
		postAttr := &PostAttr{}
		user, _ := m.db.GetUserByCookie(cookie.Value)
		singlePost.AuthUser = user
		singlePost.Post, err = m.db.GetPostByPID(int64(id))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println("Error on PostView() function")
			WarnMessage(w, "The post not found")
			return
		}

		postAttr.Comments, err = m.db.GetCommentsOfPost(int64(id))
		commentRate := model.NewCommentRating()
		for i := 0; i < len(postAttr.Comments); i++ {
			commentRate = m.db.GetRateCountOfComment(postAttr.Comments[i].CommentID, int64(id))
			postAttr.Comments[i].LikeCnt = commentRate.LikeCount
			postAttr.Comments[i].DislikeCnt = commentRate.DislikeCount
		}
		postAttr.Categories, err = m.db.GetCategoryOfPost(int64(id))
		singlePost.PostInfo = append(singlePost.PostInfo, postAttr)
		singlePost.Post.ID = int64(id)

		if r.Method == "POST" {
			r.ParseForm()
			comment := model.NewComment()
			comment.Content = r.PostFormValue("comment")
			comment.CreationDate = time.Now().Format("January 2 15:04")
			comment.PostID = int64(id)
			comment.Author = user.Username
			if comment.Content == "" || len(comment.Content) > 120 {
				singlePost.Errors = "Please write shorter comments"
				tpl.ExecuteTemplate(w, "postView.html", singlePost)
				return
			}
			if ok := m.db.AddComment(comment); !ok {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("AddComment error")
				WarnMessage(w, "Something went wrong")
				return
			}
			postID := strconv.Itoa(id)
			http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusOK)

			tpl.ExecuteTemplate(w, "postView.html", singlePost)
		}
		singlePost.Post = nil
		singlePost.PostInfo = nil
		singlePost.AuthUser = nil
	}
}

// GetAllPosts ...
func (m *Multiplexer) GetAllPosts(w http.ResponseWriter) []*model.Post {

	posts, err := m.db.GetPosts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WarnMessage(w, "Something went wrong")
		return nil
	}
	return posts
}
