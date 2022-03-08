/*
Copyright © 2021 Yale University

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// MoverCreateHandler creates a new Datasync mover
func (s *server) MoverCreateHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	group := vars["group"]

	// read the input against our struct in api/types.go
	req := DatamoverCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		msg := fmt.Sprintf("cannot decode body into create data mover input: %s", err)
		handleError(w, apierror.New(apierror.ErrBadRequest, msg, err))
		return
	}

	if req.Name == nil {
		handleError(w, apierror.New(apierror.ErrBadRequest, "Name is a required field", nil))
		return
	}

	regexName := "^[a-zA-Z0-9-]+$"
	re := regexp.MustCompile(regexName)
	if !re.MatchString(*req.Name) {
		handleError(w, apierror.New(apierror.ErrBadRequest, "Name doesn't match regex "+regexName, nil))
		return
	}

	if req.Source == nil || req.Destination == nil || req.Source.Type == "" || req.Destination.Type == "" {
		handleError(w, apierror.New(apierror.ErrBadRequest, "Source and Destination are required", nil))
		return
	}

	policy, err := s.moverCreatePolicy()
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to generate policy", err))
		return
	}

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncFullAccess",
			},
			inlinePolicy: policy,
		},
	)
	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	task, err := orch.datamoverCreate(r.Context(), group, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Flywheel-Task", task.ID)
	w.WriteHeader(http.StatusAccepted)
}

// MoverDeleteHandler deletes a Datasync mover
func (s *server) MoverDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	group := vars["group"]
	name := vars["name"]

	policy, err := s.moverDeletePolicy()
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to generate policy", err))
		return
	}

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncFullAccess",
				"arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
			inlinePolicy: policy,
		},
	)
	if err != nil {
		handleError(w, errors.Wrap(err, "unable to delete datasync orchestrator"))
		return
	}

	if err := orch.datamoverDelete(r.Context(), group, name); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MoverListHandler lists all of the data movers in a group by id
func (s *server) MoverListHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	group := vars["group"]

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncReadOnlyAccess",
				"arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
		},
	)
	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	resp, err := orch.datamoverList(r.Context(), group)
	if err != nil {
		handleError(w, err)
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to marshal json", err))
		return
	}

	w.Header().Set("X-Items", strconv.Itoa(len(resp)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

// MoverShowHandler shows details about a Datasync mover
func (s *server) MoverShowHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	group := vars["group"]
	name := vars["name"]

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncReadOnlyAccess",
				"arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
		},
	)
	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	resp, err := orch.datamoverDescribe(r.Context(), group, name)
	if err != nil {
		handleError(w, err)
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to marshal json", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
func (s *server) MoverShowrunHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	group := vars["group"]
	name := vars["name"]

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncReadOnlyAccess",
				"arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
		},
	)

	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	//Got info about the move
	respL, errL := orch.datamoverDescribe(r.Context(), group, name)
	if errL != nil {
		handleError(w, errL)
		return
	}
	// from move get info of all executions (runs) for a taskArn (moveId)
	resp, err := orch.taskRunsFromTaskArn(r.Context(), *respL.Task.TaskArn)
	if err != nil {
		handleError(w, err)
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to marshal json", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
func (s *server) MoverShowrunbyIDHandler(w http.ResponseWriter, r *http.Request) {
	w = LogWriter{w}
	vars := mux.Vars(r)
	account := vars["account"]
	// group := vars["group"]
	// name := vars["name"]
	id := vars["id"]
	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncReadOnlyAccess",
				"arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
		},
	)

	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	// //Got info about the move
	// respL, errL := orch.datamoverDescribe(r.Context(), group, name)
	// if errL != nil {
	// 	handleError(w, errL)
	// 	return
	// }

	// from move get info of all executions (runs) for a taskArn (moveId)
	resp, err := orch.TaskDetailsFromid(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	j, err := json.Marshal(resp)
	if err != nil {
		handleError(w, apierror.New(apierror.ErrInternalError, "failed to marshal json", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
