package service

import (
	"errors"
	"github.com/google/uuid"
	"noda/api/data/model"
	"noda/api/data/transfer"
	"noda/api/repository"
	"noda/failure"
	"strings"
)

type ListService struct {
	r repository.IListRepository
}

func NewListService(r repository.IListRepository) *ListService {
	return &ListService{r}
}

func (s *ListService) SaveList(ownerID, groupID uuid.UUID, next *transfer.ListCreation) (insertedID uuid.UUID, err error) {
	switch {
	case uuid.Nil == ownerID:
		return uuid.Nil, failure.NewNilParameterError("SaveList", "ownerID")
	case uuid.Nil == groupID:
		return uuid.Nil, failure.NewNilParameterError("SaveList", "groupID")
	case nil == next:
		return uuid.Nil, failure.NewNilParameterError("SaveList", "next")
	}
	next.Name = strings.Trim(next.Name, " \t\n")
	next.Description = strings.Trim(next.Description, " \t\n")
	switch {
	case "" == next.Name:
		return uuid.Nil, errors.New("name cannot be an empty string")
	case 50 < len(next.Name):
		return uuid.Nil, errors.New("name too long for list: max length must be 50")
	}
	id, err := s.r.InsertList(ownerID.String(), groupID.String(), next)
	if nil != err {
		return uuid.Nil, err
	}
	return uuid.Parse(id)
}

func (s *ListService) FindListByID(ownerID, groupID, listID uuid.UUID) (list *model.List, err error) {
	switch {
	case uuid.Nil == ownerID:
		return nil, failure.NewNilParameterError("FindListByID", "ownerID")
	case uuid.Nil == groupID:
		return nil, failure.NewNilParameterError("FindListByID", "groupID")
	case uuid.Nil == listID:
		return nil, failure.NewNilParameterError("FindListByID", "listID")
	}
	return s.r.FetchListByID(ownerID.String(), groupID.String(), listID.String())
}
