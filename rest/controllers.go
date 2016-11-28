package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/chmouel/chmoufrack"
	"github.com/gorilla/mux"
)

func GETPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := chmoufrack.GetPrograms()
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(w).Encode(programs); err != nil {
		panic(err)
	}
}

func CreateProgram(writer http.ResponseWriter, reader *http.Request) {
	var program chmoufrack.Program
	if reader.Body == nil {
		http.Error(writer, "Please send a request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(reader.Body).Decode(&program)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := chmoufrack.CreateProgram(program.Name, program.Comment); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(http.StatusCreated)
}

func CreateMultipleWorkouts(writer http.ResponseWriter, reader *http.Request) {
	var workouts []chmoufrack.Workout
	if reader.Body == nil {
		http.Error(writer, "Please send a request body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(reader.Body).Decode(&workouts)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	for _, workout := range workouts {
		if err = convertAndCreateWorkout(workout); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}
	}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(http.StatusCreated)
}

//RESTProgramCleanup Cleanup all workout of a program
func CleanupProgram(writer http.ResponseWriter, reader *http.Request) {
	vars := mux.Vars(reader)
	programName := vars["name"]

	p, err := chmoufrack.GetProgram(programName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = chmoufrack.DeleteAllWorkoutProgram(p.ID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(http.StatusCreated)
}

func convertAndCreateWorkout(w chmoufrack.Workout) (err error) {
	var percentage, meters, repetition int

	p, err := chmoufrack.GetProgram(w.ProgramName)
	if p.ID == 0 {
		return errors.New("Cannot find programName " + w.ProgramName)
	}

	if repetition, err = strconv.Atoi(w.Repetition); err != nil {
		return
	}

	if meters, err = strconv.Atoi(w.Meters); err != nil {
		return
	}

	if percentage, err = strconv.Atoi(w.Percentage); err != nil {
		return
	}

	_, err = chmoufrack.CreateWorkout(repetition, meters, percentage, w.Repos, int(p.ID))
	return
}