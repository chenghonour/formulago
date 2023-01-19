/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"fmt"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/dictionary"
	"formulago/data/ent/dictionarydetail"
	"formulago/data/ent/predicate"
	"github.com/cockroachdb/errors"
	"time"
)

type Dictionary struct {
	Data *data.Data
}

func NewDictionary(data *data.Data) domain.Dictionary {
	return &Dictionary{
		Data: data,
	}
}

func (d *Dictionary) Create(ctx context.Context, req *domain.DictionaryInfo) error {
	// whether dictionary name exists
	dictionaryExist, _ := d.Data.DBClient.Dictionary.Query().Where(dictionary.Name(req.Name)).Exist(ctx)
	if dictionaryExist {
		return errors.New("dictionary name already exists")
	}
	// create dictionary
	_, err := d.Data.DBClient.Dictionary.Create().
		SetTitle(req.Title).
		SetName(req.Name).
		SetStatus(uint8(req.Status)).
		SetDescription(req.Description).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "create Dictionary failed")
	}
	return nil
}

func (d *Dictionary) Update(ctx context.Context, req *domain.DictionaryInfo) error {
	// whether dictionary is exists
	dictionaryExist, _ := d.Data.DBClient.Dictionary.Query().Where(dictionary.ID(req.ID)).Exist(ctx)
	if !dictionaryExist {
		return errors.New("The dictionary try to update is not exists")
	}
	// update dictionary
	_, err := d.Data.DBClient.Dictionary.UpdateOneID(req.ID).
		SetTitle(req.Title).
		SetName(req.Name).
		SetStatus(uint8(req.Status)).
		SetDescription(req.Description).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "update Dictionary failed")
	}
	return nil
}

func (d *Dictionary) Delete(ctx context.Context, id uint64) error {
	// delete dictionary
	err := d.Data.DBClient.Dictionary.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete Dictionary failed")
	}
	return nil
}

func (d *Dictionary) List(ctx context.Context, req *domain.DictListReq) (list []*domain.DictionaryInfo, total int, err error) {
	// query dictionary
	var predicates []predicate.Dictionary
	if req.Title != "" {
		predicates = append(predicates, dictionary.TitleContains(req.Title))
	}
	if req.Name != "" {
		predicates = append(predicates, dictionary.NameContains(req.Name))
	}
	dictionaries, err := d.Data.DBClient.Dictionary.Query().Where(predicates...).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query Dictionary list failed")
	}

	// format result
	for _, dict := range dictionaries {
		list = append(list, &domain.DictionaryInfo{
			ID:          dict.ID,
			Title:       dict.Title,
			Name:        dict.Name,
			Status:      uint64(dict.Status),
			Description: dict.Description,
			CreatedAt:   dict.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   dict.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	total, _ = d.Data.DBClient.Dictionary.Query().Where(predicates...).Count(ctx)
	return
}

func (d *Dictionary) CreateDetail(ctx context.Context, req *domain.DictionaryDetail) error {
	// whether dictionary detail is exists
	exist, err := d.Data.DBClient.DictionaryDetail.Query().Where(dictionarydetail.Key(req.Key)).
		Where(dictionarydetail.Value(req.Value)).Where(dictionarydetail.HasDictionaryWith(dictionary.ID(req.ParentID))).Exist(ctx)
	if err != nil {
		return errors.Wrap(err, "query DictionaryDetail exist failed")
	}
	if exist {
		return errors.New("dictionary detail already exists")
	}

	// find dictionary by id
	dict, err := d.Data.DBClient.Dictionary.Query().Where(dictionary.ID(req.ParentID)).Only(ctx)
	if err != nil {
		return errors.Wrap(err, "query Dictionary failed")
	}
	if dict == nil {
		return errors.New(fmt.Sprintf("dictionary not found, please check dictionary id, %d", req.ParentID))
	}

	// create dictionary detail
	_, err = d.Data.DBClient.DictionaryDetail.Create().
		SetDictionary(dict). // set parent dictionary
		SetTitle(req.Title).
		SetKey(req.Key).
		SetValue(req.Value).
		SetStatus(uint8(req.Status)).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "create DictionaryDetail failed")
	}
	return nil
}

func (d *Dictionary) UpdateDetail(ctx context.Context, req *domain.DictionaryDetail) error {
	// query dictionary detail
	detail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.ID(req.ID)).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).Only(ctx)
	if err != nil {
		return errors.Wrap(err, "query DictionaryDetail failed")
	}
	// update dictionary detail
	_, err = d.Data.DBClient.DictionaryDetail.UpdateOneID(req.ID).
		SetTitle(req.Title).
		SetKey(req.Key).
		SetValue(req.Value).
		SetStatus(uint8(req.Status)).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "update DictionaryDetail failed")
	}
	// delete dictionary detail from cache
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-%s", detail.Edges.Dictionary.Name, detail.Key))
	return nil
}

func (d *Dictionary) DeleteDetail(ctx context.Context, id uint64) error {
	// query dictionary detail
	detail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.ID(id)).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).Only(ctx)
	if err != nil {
		return errors.Wrap(err, "query DictionaryDetail failed")
	}
	// delete dictionary detail
	err = d.Data.DBClient.DictionaryDetail.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete DictionaryDetail failed")
	}
	// delete dictionary detail from cache
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-%s", detail.Edges.Dictionary.Name, detail.Key))
	return nil
}

func (d *Dictionary) DetailListByDictName(ctx context.Context, dictName string) (list []*domain.DictionaryDetail, total uint64, err error) {
	// query dictionary detail
	details, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query DictionaryDetail list failed")
	}

	// format result
	for _, detail := range details {
		list = append(list, &domain.DictionaryDetail{
			ID:        detail.ID,
			Title:     detail.Title,
			Key:       detail.Key,
			Value:     detail.Value,
			Status:    uint64(detail.Status),
			CreatedAt: detail.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: detail.UpdatedAt.Format("2006-01-02 15:04:05"),
			ParentID:  detail.Edges.Dictionary.ID,
		})
	}
	total = uint64(len(list))
	return
}

func (d *Dictionary) DetailByDictNameAndKey(ctx context.Context, dictName, key string) (detail *domain.DictionaryDetail, err error) {
	// query dictionary detail from cache
	v, found := d.Data.Cache.Get(fmt.Sprintf("Dictionary%s-%s", dictName, key))
	if found {
		return v.(*domain.DictionaryDetail), nil
	}
	// query dictionary detail from database
	dictDetail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		Where(dictionarydetail.Key(key)).First(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query DictionaryDetail list failed")
	}

	// format result
	detail = new(domain.DictionaryDetail)
	detail.ID = dictDetail.ID
	detail.Title = dictDetail.Title
	detail.Key = dictDetail.Key
	detail.Value = dictDetail.Value
	detail.Status = uint64(dictDetail.Status)
	detail.CreatedAt = dictDetail.CreatedAt.Format("2006-01-02 15:04:05")
	detail.UpdatedAt = dictDetail.UpdatedAt.Format("2006-01-02 15:04:05")

	// set cache
	d.Data.Cache.Set(fmt.Sprintf("Dictionary%s-%s", dictName, key), detail, 24*time.Hour)

	return detail, nil
}
