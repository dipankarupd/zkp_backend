package routes

import (
	"github.com/dipankarupd/zkp/app/controllers"
	"github.com/gorilla/mux"
)

var RegisterNewRoutes = func(router *mux.Router) {
	router.HandleFunc("/register", controllers.RegisterUser).Methods("POST")
	router.HandleFunc("/user/{token}", controllers.GetUserByToken).Methods("GET")
	router.HandleFunc("/students/{token}/attendance", controllers.UpdateAttendance).Methods("PUT")

}
