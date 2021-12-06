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
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

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
		handleError(w, errors.Wrap(err, "failed to list data movers"))
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
	id := vars["id"]

	orch, err := s.newDatasyncOrchestrator(
		r.Context(),
		account,
		&sessionParams{
			role: fmt.Sprintf("arn:aws:iam::%s:role/%s", account, s.session.RoleName),
			policyArns: []string{
				"arn:aws:iam::aws:policy/AWSDataSyncReadOnlyAccess",
				// "arn:aws:iam::aws:policy/ResourceGroupsandTagEditorReadOnlyAccess",
			},
		},
	)
	if err != nil {
		handleError(w, errors.Wrap(err, "unable to create datasync orchestrator"))
		return
	}

	resp, err := orch.datamoverDescribe(r.Context(), group, id)
	if err != nil {
		handleError(w, errors.Wrap(err, "failed to describe data mover"))
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
