// Package userapi reads Riido with a standalone authenticated user session.
package userapi

import (
	"context"
	"errors"
	"net/http"
)

const maxBodyBytes = 2 << 20

var (
	ErrInvalidResponse = errors.New("riido API returned an invalid response")
	ErrAuthentication  = errors.New("riido CLI authentication is required")
	ErrNotFound        = errors.New("the requested resource was not found")
	ErrPermission      = errors.New("the current user cannot read this resource")
	ErrRateLimited     = errors.New("riido API read rate limit was exceeded")
	ErrUnavailable     = errors.New("riido API is temporarily unavailable")
	ErrGraphQL         = errors.New("riido could not complete this read")
	ErrConflict        = errors.New("the requested resource version has changed")
	ErrInvalidRequest  = errors.New("the read request parameters are invalid")
)

type ReadClient interface {
	DoRead(context.Context, *http.Request) (*http.Response, error)
}
type Reader struct {
	Endpoint string
	Client   ReadClient
}
type Identity struct {
	ID                    string `json:"id"`
	Email                 string `json:"email,omitempty"`
	Username              string `json:"username,omitempty"`
	LoginType             string `json:"loginType,omitempty"`
	LastViewedWorkspaceID string `json:"lastViewedWorkspaceId,omitempty"`
}
type Work struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	ComponentType      string       `json:"componentType,omitempty"`
	KeyNumber          string       `json:"keyNumber,omitempty"`
	Description        string       `json:"description,omitempty"`
	CreatedAt          string       `json:"createdAt,omitempty"`
	UpdatedAt          string       `json:"updatedAt,omitempty"`
	Document           *Document    `json:"document,omitempty"`
	StartDate          *string      `json:"startDate"`
	EndDate            *string      `json:"endDate"`
	Estimate           *int         `json:"estimate"`
	TotalPriorityScore *int         `json:"totalPriorityScore"`
	IsDone             bool         `json:"isDone"`
	IsArchived         bool         `json:"isArchived"`
	Status             *WorkStatus  `json:"status"`
	ComponentMembers   []WorkMember `json:"componentMembers"`
	ComponentLabels    []WorkLabel  `json:"componentLabels"`
}
type Document struct {
	ID               string `json:"id"`
	HTMLContent      string `json:"HTMLContent"`
	TiptapDocumentID string `json:"tiptapDocumentId,omitempty"`
}
