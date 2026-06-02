/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"fmt"
"errors"
	"formulago/biz/domain/admin"
	"formulago/pkg/times"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/dictionary"
	"formulago/data/ent/dictionarydetail"
	"formulago/data/ent/predicate"
	"time"
)

type Dictionary struct {
	Data *data.Data
}

func NewDictionary(data *data.Data) admin.Dictionary {
	return &Dictionary{
		Data: data,
	}
}

func (d *Dictionary) Create(ctx context.Context, req *admin.DictionaryInfo) error {
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
		return fmt.Errorf("create Dictionary failed: %w", err)
	}
	return nil
}

func (d *Dictionary) Update(ctx context.Context, req *admin.DictionaryInfo) error {
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
		return fmt.Errorf("update Dictionary failed: %w", err)
	}
	return nil
}

func (d *Dictionary) Delete(ctx context.Context, id uint64) error {
	// whether dictionary is exists
	dict, err := d.Data.DBClient.Dictionary.Query().Where(dictionary.ID(id)).Only(ctx)
	if err != nil {
		return fmt.Errorf("query Dictionary failed: %w", err)
	}
	if dict == nil {
		return fmt.Errorf("The dictionary(id=%d) try to delete is not exists", id)
	}
	// whether dictionary has detail
	// query dictionary detail
	details, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dict.Name))).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).All(ctx)
	if err != nil {
		return fmt.Errorf("query DictionaryDetail failed: %w", err)
	}
	if len(details) > 0 {
		return errors.New("dictionary has detail, please delete detail first")
	}
	// delete dictionary
	err = d.Data.DBClient.Dictionary.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete Dictionary failed: %w", err)
	}
	return nil
}

func (d *Dictionary) List(ctx context.Context, req *admin.DictListReq) (list []*admin.DictionaryInfo, total int, err error) {
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
		return nil, 0, fmt.Errorf("query Dictionary list failed: %w", err)
	}

	// format result
	for _, dict := range dictionaries {
		list = append(list, &admin.DictionaryInfo{
			ID:          dict.ID,
			Title:       dict.Title,
			Name:        dict.Name,
			Status:      uint64(dict.Status),
			Description: dict.Description,
			CreatedAt:   dict.CreatedAt.Format(times.TimeFormat),
			UpdatedAt:   dict.UpdatedAt.Format(times.TimeFormat),
		})
	}
	total, err = d.Data.DBClient.Dictionary.Query().Where(predicates...).Count(ctx)
	if err != nil {
		err = fmt.Errorf("count dictionary failed: %w", err)
		return
	}
	return
}

func (d *Dictionary) CreateDetail(ctx context.Context, req *admin.DictionaryDetail) error {
	// whether dictionary detail is exists
	exist, err := d.Data.DBClient.DictionaryDetail.Query().Where(dictionarydetail.Key(req.Key)).
		Where(dictionarydetail.Value(req.Value)).Where(dictionarydetail.HasDictionaryWith(dictionary.ID(req.ParentID))).Exist(ctx)
	if err != nil {
		return fmt.Errorf("query DictionaryDetail exist failed: %w", err)
	}
	if exist {
		return errors.New("dictionary detail already exists")
	}

	// find dictionary by id
	dict, err := d.Data.DBClient.Dictionary.Query().Where(dictionary.ID(req.ParentID)).Only(ctx)
	if err != nil {
		return fmt.Errorf("query Dictionary failed: %w", err)
	}
	if dict == nil {
		return fmt.Errorf("dictionary not found, please check dictionary id, %d", req.ParentID)
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
		return fmt.Errorf("create DictionaryDetail failed: %w", err)
	}
	return nil
}

func (d *Dictionary) UpdateDetail(ctx context.Context, req *admin.DictionaryDetail) error {
	// query dictionary detail
	detail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.ID(req.ID)).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).Only(ctx)
	if err != nil {
		return fmt.Errorf("query DictionaryDetail failed: %w", err)
	}
	// update dictionary detail
	_, err = d.Data.DBClient.DictionaryDetail.UpdateOneID(req.ID).
		SetTitle(req.Title).
		SetKey(req.Key).
		SetValue(req.Value).
		SetStatus(uint8(req.Status)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update DictionaryDetail failed: %w", err)
	}
	// delete dictionary detail from cache
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-key%s", detail.Edges.Dictionary.Name, detail.Key))
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-value%s", detail.Edges.Dictionary.Name, detail.Value))
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
		return fmt.Errorf("query DictionaryDetail failed: %w", err)
	}
	// delete dictionary detail
	err = d.Data.DBClient.DictionaryDetail.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete DictionaryDetail failed: %w", err)
	}
	// delete dictionary detail from cache
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-key%s", detail.Edges.Dictionary.Name, detail.Key))
	d.Data.Cache.Delete(fmt.Sprintf("Dictionary%s-value%s", detail.Edges.Dictionary.Name, detail.Value))
	return nil
}

func (d *Dictionary) DetailListByDictName(ctx context.Context, dictName string) (list []*admin.DictionaryDetail, total uint64, err error) {
	// query dictionary detail
	details, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query DictionaryDetail list failed: %w", err)
	}

	// format result
	for _, detail := range details {
		list = append(list, &admin.DictionaryDetail{
			ID:        detail.ID,
			Title:     detail.Title,
			Key:       detail.Key,
			Value:     detail.Value,
			Status:    uint64(detail.Status),
			CreatedAt: detail.CreatedAt.Format(times.TimeFormat),
			UpdatedAt: detail.UpdatedAt.Format(times.TimeFormat),
			ParentID:  detail.Edges.Dictionary.ID,
		})
	}
	total = uint64(len(list))
	return
}

func (d *Dictionary) K2VMapByDictName(ctx context.Context, dictName string) (K2VMap map[string]string, err error) {
	K2VMap = make(map[string]string)
	// query dictionary detail
	details, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query DictionaryDetail list failed: %w", err)
	}

	// format result
	for _, detail := range details {
		K2VMap[detail.Key] = detail.Value
	}
	return K2VMap, nil
}

func (d *Dictionary) V2KMapByDictName(ctx context.Context, dictName string) (V2KMap map[string]string, err error) {
	V2KMap = make(map[string]string)
	// query dictionary detail
	details, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		// union query to get the fields of the associated table
		WithDictionary(func(q *ent.DictionaryQuery) {
			// get all fields default, or use q.Select() to get some fields
		}).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query DictionaryDetail list failed: %w", err)
	}

	// format result
	for _, detail := range details {
		V2KMap[detail.Value] = detail.Key
	}
	return V2KMap, nil
}

func (d *Dictionary) DetailByDictNameAndKey(ctx context.Context, dictName, key string) (detail *admin.DictionaryDetail, err error) {
	// query dictionary detail from cache
	v, found := d.Data.Cache.Get(fmt.Sprintf("Dictionary%s-key%s", dictName, key))
	if found {
		return v.(*admin.DictionaryDetail), nil
	}
	// query dictionary detail from database
	dictDetail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		Where(dictionarydetail.Key(key)).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("query DictionaryDetail list failed: %w", err)
	}

	// format result
	detail = new(admin.DictionaryDetail)
	detail.ID = dictDetail.ID
	detail.Title = dictDetail.Title
	detail.Key = dictDetail.Key
	detail.Value = dictDetail.Value
	detail.Status = uint64(dictDetail.Status)
	detail.CreatedAt = dictDetail.CreatedAt.Format(times.TimeFormat)
	detail.UpdatedAt = dictDetail.UpdatedAt.Format(times.TimeFormat)

	// set cache
	d.Data.Cache.Set(fmt.Sprintf("Dictionary%s-key%s", dictName, key), detail, 72*time.Hour)
	return detail, nil
}

func (d *Dictionary) DetailByDictNameAndValue(ctx context.Context, dictName, value string) (detail *admin.DictionaryDetail, err error) {
	// query dictionary detail from cache
	v, found := d.Data.Cache.Get(fmt.Sprintf("Dictionary%s-value%s", dictName, value))
	if found {
		return v.(*admin.DictionaryDetail), nil
	}
	// query dictionary detail from database
	dictDetail, err := d.Data.DBClient.DictionaryDetail.Query().
		Where(dictionarydetail.HasDictionaryWith(dictionary.NameEQ(dictName))).
		Where(dictionarydetail.Value(value)).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("query DictionaryDetail list failed: %w", err)
	}

	// format result
	detail = new(admin.DictionaryDetail)
	detail.ID = dictDetail.ID
	detail.Title = dictDetail.Title
	detail.Key = dictDetail.Key
	detail.Value = dictDetail.Value
	detail.Status = uint64(dictDetail.Status)
	detail.CreatedAt = dictDetail.CreatedAt.Format(times.TimeFormat)
	detail.UpdatedAt = dictDetail.UpdatedAt.Format(times.TimeFormat)

	// set cache
	d.Data.Cache.Set(fmt.Sprintf("Dictionary%s-value%s", dictName, value), detail, 72*time.Hour)
	return detail, nil
}
