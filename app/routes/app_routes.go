package routes

import (
	"github.com/dipankarupd/zkp/app/controllers"
	"github.com/dipankarupd/zkp/app/db"
	"github.com/gorilla/mux"
)

func RegisterNewRoutes(router *mux.Router, database db.DB) {

	// router.HandleFunc("/register", controllers.RegisterUser).Methods("POST")
	// router.HandleFunc("/user/{token}", controllers.GetUserByToken).Methods("GET")
	// router.HandleFunc("/students/{token}/attendance", controllers.UpdateAttendance).Methods("PUT")

	/// Student Routes:
	router.HandleFunc("/api/student/register", controllers.RegisterUser).Methods("POST")
	router.HandleFunc("/api/student/user/{token}", controllers.GetUserByToken).Methods("GET")
	router.HandleFunc("/api/classrooms/{id}/join", controllers.JoinClassroom).Methods("POST")
	router.HandleFunc("/api/student/{token}/classes/{id}/attendance", controllers.MarkAttendance).Methods("POST")
	router.HandleFunc("/api/students/{token}/classrooms/{classroom_id}/progress", controllers.ViewHistory).Methods("GET")

	/// Teacher Routes:
	router.HandleFunc("/api/classrooms/{id}/classes", controllers.CreateClass).Methods("POST")
	router.HandleFunc("/api/classrooms", controllers.CreateClassroom).Methods("POST")
	router.HandleFunc("/api/classrooms", controllers.GetAllClassrooms).Methods("GET")
	router.HandleFunc("/api/classrooms/{id}", controllers.GetClassroomDetails).Methods("GET")
	// router.HandleFunc("/api/classrooms/{id}/students", controllers.GetAllStudents).Methods("GET")
	router.HandleFunc("/api/classrooms/{id}/students/{student_id}/attendance", controllers.GetIndividualStudent).Methods("GET")
	router.HandleFunc("/api/classrooms/{id}/classes", controllers.GetAllClasses).Methods("GET")
	router.HandleFunc("/api/classrooms/{id}/export", controllers.MakeCSV).Methods("GET")

}
