package main

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/samarabbas/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence"
)

/**
 * Dummy server that support to list expenses, create new expense, update expense state and checking expense state.
 */

type expenseState string

const (
	created   expenseState = "CREATED"
	approved               = "APPROVED"
	rejected               = "REJECTED"
	completed              = "COMPLETED"
)

// use memory store for this dummy server
var allExpense = make(map[string]expenseState)

var tokenMap = make(map[string][]byte)

var workflowClient cadence.Client

func main() {
	var h common.SampleHelper
	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting dummy server...")
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/action", actionHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/registerCallback", callbackHandler)
	http.ListenAndServe(":8080", nil)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>DUMMY EXPENSE SYSTEM</h1>"+"<a href=\"/list\">HOME</a>"+
		"<h3>All expense requests:</h3><table border=1><tr><th>Expense ID</th><th>Status</th><th>Action</th>")
	keys := []string{}
	for k := range allExpense {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		state := allExpense[id]
		actionLink := ""
		if state == created {
			actionLink = fmt.Sprintf("<a href=\"/action?type=approve&id=%s\">"+
				"<button style=\"background-color:#4CAF50;\">APPROVE</button></a>"+
				"&nbsp;&nbsp;<a href=\"/action?type=reject&id=%s\">"+
				"<button style=\"background-color:#f44336;\">REJECT</button></a>", id, id)
		}
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>", id, state, actionLink)
	}
	fmt.Fprint(w, "</table>")
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	oldState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	actionType := r.URL.Query().Get("type")
	switch actionType {
	case "approve":
		allExpense[id] = approved
	case "reject":
		allExpense[id] = rejected
	case "payment":
		allExpense[id] = completed
	}
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		listHandler(w, r)
	}

	if oldState == created && (allExpense[id] == approved || allExpense[id] == rejected) {
		// report state change
		notifyExpenseStateChange(id, string(allExpense[id]))
	}

	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allExpense[id])
	return
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	_, ok := allExpense[id]
	if ok {
		fmt.Fprint(w, "ERROR:ID_ALREADY_EXISTS")
		return
	}

	allExpense[id] = created
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		listHandler(w, r)
	}
	fmt.Printf("Created new expense id:%s.\n", id)
	return
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	state, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}

	fmt.Fprint(w, state)
	fmt.Printf("Checking status for %s: %s\n", id, state)
	return
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	currState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if currState != created {
		fmt.Fprint(w, "ERROR:INVALID_STATE")
		return
	}

	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
		fmt.Fprint(w, "ERROR:INVALID_FORM_DATA")
		return
	}

	taskToken := r.PostFormValue("task_token")
	fmt.Printf("Registered callback for ID=%s, token=%s\n", id, taskToken)
	tokenMap[id] = []byte(taskToken)
	fmt.Fprint(w, "SUCCEED")
}

func notifyExpenseStateChange(id, state string) {
	token, ok := tokenMap[id]
	if !ok {
		fmt.Printf("Invalid id:%s\n", id)
		return
	}
	err := workflowClient.CompleteActivity(token, state, nil)
	if err != nil {
		fmt.Printf("Failed to complete activity with error: %+v\n", err)
	} else {
		fmt.Printf("Successfully complete activity: %s\n", token)
	}
}
