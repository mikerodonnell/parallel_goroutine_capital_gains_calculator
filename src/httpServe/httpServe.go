package httpServe

import (
	"dba"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const authRepIdKey = "authRepId"

func Serve() {
	fmt.Println("~~~~~~~~~~~~~~~~ starting server")

	getUserHandler := func(writer http.ResponseWriter, request *http.Request) {
		params := request.URL.Query()
		authRepIds, ok := params[authRepIdKey]
		if !ok {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(writer, "parameter %s is required", authRepIdKey)

			return
		}

		authRepId := authRepIds[0]
		rep, err := dba.GetRep(authRepId)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(writer, "error lookup up rep %s", authRepId)

			return
		}

		if rep == nil {
			writer.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(writer, "rep %s does not exist", authRepIdKey)

			return
		}

		fmt.Fprintf(writer, "rep details: %s", rep)

		return
	}

	createUserHandler := func(writer http.ResponseWriter, request *http.Request) {
		decoder := json.NewDecoder(request.Body)

		var transientRep dba.Rep

		err := decoder.Decode(&transientRep)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(writer, "invalid rep request body %s", request.Body)

			return
		}

		persistentRepId, err := dba.CreateRep(transientRep.AuthRepId)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(writer, "error creating rep %s", transientRep.AuthRepId)

			return
		}

		writer.WriteHeader(http.StatusCreated)
		fmt.Fprintf(writer, "created rep %d", persistentRepId)

		return
	}

	userHandler := func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			getUserHandler(writer, request)
		case http.MethodPut:
			createUserHandler(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(writer, "HTTP method %s is not supported for /users", request.Method)
		}

		return
	}

	http.HandleFunc("/users", userHandler)

	err := http.ListenAndServe(":7000", nil)
	if err != nil {
		log.Fatal(err)
	}

	return
}
