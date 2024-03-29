package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func SetOnline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	app := ps.ByName("app")
	ip := ps.ByName("ip")
	if OnlineStater.Set(app, ip) {
		fmt.Fprint(w, "ok")
		return
	}
	fmt.Fprint(w, "not ok")
}

func GetOnline(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	count := OnlineStater.Get(ps.ByName("app"))
	fmt.Fprint(w, count)
}

func GetOnlineDump(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ips := OnlineStater.Dump(ps.ByName("app"))
	jsonIps, _ := json.Marshal(ips)
	fmt.Fprint(w, string(jsonIps))
}

func RegistRoutes(r *httprouter.Router) {
	r.POST("/online/:app/:ip", SetOnline)
	r.GET("/online/:app", GetOnline)
	r.GET("/online/:app/dump", GetOnlineDump)
}
