package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

var access string

func getAccess() bool {
	return access != ""
}

type User struct {
	Name        string     `json:"name"`
	Avatar      string     `json:"avatar"`
	UserID      string     `json:"userid"`
	Email       string     `json:"email"`
	Signature   string     `json:"signature"`
	Title       string     `json:"title"`
	Group       string     `json:"group"`
	Tags        []Tag      `json:"tags"`
	NotifyCount int        `json:"notifyCount"`
	UnreadCount int        `json:"unreadCount"`
	Country     string     `json:"country"`
	Access      string     `json:"access"`
	Geographic  Geographic `json:"geographic"`
	Address     string     `json:"address"`
	Phone       string     `json:"phone"`
}

type Geographic struct {
	Provice Tag `json:"province"`
	City    Tag `json:"city"`
}
type Tag struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type UserInfo struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

type Error struct {
	Data         map[string]interface{} `json:"data"`
	ErrorCode    string                 `json:"errorCode"`
	ErrorMessage string                 `json:"errorMessage"`
	Success      bool                   `json:"access"`
}

func CurrentUser(w http.ResponseWriter, r *http.Request) {
	if !getAccess() {
		send(w, http.StatusForbidden, Error{
			Data: map[string]interface{}{
				"isLogin": false,
			},
			ErrorCode:    "401",
			ErrorMessage: "请先登录！",
			Success:      true,
		})
		return
	}
	send(w, http.StatusOK, User{
		Name:      "Serati Ma",
		Avatar:    "https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png",
		UserID:    "00000001",
		Email:     "antdesign@alipay.com",
		Signature: "海纳百川，有容乃大",
		Title:     "交互专家",
		Group:     "蚂蚁金服－某某某事业群－某某平台部－某某技术部－UED",
		Tags: []Tag{
			{
				Key:   "0",
				Label: "很有想法的",
			},
			{
				Key:   "1",
				Label: "专注设计",
			},
			{
				Key:   "2",
				Label: "辣~",
			},
			{
				Key:   "3",
				Label: "大长腿",
			},
			{
				Key:   "4",
				Label: "川妹子",
			},
			{
				Key:   "5",
				Label: "海纳百川",
			},
		},
		NotifyCount: 12,
		UnreadCount: 11,
		Country:     "China",
		Access:      access,
		Geographic: Geographic{
			Provice: Tag{
				Label: "浙江省",
				Key:   "330000",
			},
			City: Tag{
				Label: "杭州市",
				Key:   "330100",
			},
		},
		Address: "西湖区工专路 77 号",
		Phone:   "0752-268888888",
	})
}

func send(w http.ResponseWriter, code int, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(o)
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusOK, []UserInfo{
		{
			Key:     "1",
			Name:    "John Brown",
			Age:     32,
			Address: "New York No. 1 Lake Park",
		},
		{
			Key:     "2",
			Name:    "Jim Green",
			Age:     42,
			Address: "London No. 1 Lake Park",
		},
		{
			Key:     "3",
			Name:    "Joe Black",
			Age:     32,
			Address: "Sidney No. 1 Lake Park",
		},
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var o LoginRequest
	json.NewDecoder(r.Body).Decode(&o)
	if o.Password == "ant.design" && o.Username == "admin" {
		send(w, http.StatusOK, map[string]interface{}{
			"status":           "ok",
			"type":             o.Type,
			"currentAuthority": "admin",
		})
		access = "admin"
		return
	}
	if o.Password == "ant.design" && o.Username == "user" {
		send(w, http.StatusOK, map[string]interface{}{
			"status":           "ok",
			"type":             o.Type,
			"currentAuthority": "user",
		})
		access = "user"
		return
	}
	if o.Type == "mobile" {
		send(w, http.StatusOK, map[string]interface{}{
			"status":           "ok",
			"type":             o.Type,
			"currentAuthority": "admin",
		})
		access = "admin"
		return
	}
	send(w, http.StatusOK, map[string]interface{}{
		"status":           "error",
		"type":             o.Type,
		"currentAuthority": "guest",
	})
	access = "guest"
}

func Logout(w http.ResponseWriter, r *http.Request) {
	access = ""
	send(w, http.StatusOK, map[string]interface{}{"success": true})
}

func Register(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusOK, map[string]interface{}{
		"status":           "ok",
		"currentAuthority": "user",
		"success":          true,
	})
}

func E500(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusInternalServerError, map[string]interface{}{
		"timestamp": 1513932555104,
		"status":    500,
		"error":     "error",
		"message":   "error",
		"path":      "/base/category/list",
	})
}
func E400(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusNotFound, map[string]interface{}{
		"timestamp": 1513932643431,
		"status":    404,
		"error":     "Not Found",
		"message":   "No message available",
		"path":      "/base/category/list/2121212",
	})
}
func E403(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusForbidden, map[string]interface{}{
		"timestamp": 1513932555104,
		"status":    403,
		"error":     "Forbidden",
		"message":   "Forbidden",
		"path":      "/base/category/list",
	})
}

func E401(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusUnauthorized, map[string]interface{}{
		"timestamp": 1513932555104,
		"status":    403,
		"error":     "Unauthorized",
		"message":   "Unauthorized",
		"path":      "/base/category/list",
	})
}

func Captcha(w http.ResponseWriter, r *http.Request) {
	send(w, http.StatusOK, "captcha-xxx")
}

func AddRoutes(m *mux.Router) {
	m.Use(dumpRequest)
	m.HandleFunc("/api/currentUser", CurrentUser).Methods(http.MethodGet)
	m.HandleFunc("/api/users", ListUsers).Methods(http.MethodGet)
	m.HandleFunc("/api/login/account", Login).Methods(http.MethodPost)
	m.HandleFunc("/api/login/outLogin", Logout).Methods(http.MethodPost)
	m.HandleFunc("/api/register", Register).Methods(http.MethodPost)
	m.HandleFunc("/api/500", E500).Methods(http.MethodGet)
	m.HandleFunc("/api/400", E400).Methods(http.MethodGet)
	m.HandleFunc("/api/401", E401).Methods(http.MethodGet)
	m.HandleFunc("/api/403", E403).Methods(http.MethodGet)
	m.HandleFunc("/api/403", E403).Methods(http.MethodGet)
	m.HandleFunc("/api/login/captcha", Captcha).Methods(http.MethodGet)
	m.HandleFunc("/api/notices", GetNotice).Methods(http.MethodGet)
	m.HandleFunc("/api/rule", GetRule).Methods(http.MethodGet)
}

func dumpRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(b))
		h.ServeHTTP(w, r)
	})
}
