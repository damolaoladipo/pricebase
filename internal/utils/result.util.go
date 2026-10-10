package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	commonutils "github.com/damolaoladipo/common-go/utils"
)

const (
	earthRadiusKM = 6378.1
	earthRadiusML = 3963.2
)

func Advanced(ctx context.Context, pool *pgxpool.Pool, q AdvancedQuery) (*PaginatedResult, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	removeKeys := []string{"select", "sort", "page", "limit", "order", "from", "to"}
	cleanParams := make(map[string][]string)
	for k, v := range q.QueryParam {
		shouldRemove := false
		for _, rk := range removeKeys {
			if k == rk {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			cleanParams[k] = v
		}
	}

	var args []any
	var whereParts []string

	if q.Ref != "" && q.Value != nil {
		args = append(args, q.Value)
		refCol := q.Ref
		if refCol == "id" {
			refCol = "id"
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", refCol, len(args)))
	}

	if len(q.Data) > 0 {
		for _, p := range q.Data {
			args = append(args, p.Value)
			paramIdx := len(args)
			switch p.Operator {
			case "gt":
				whereParts = append(whereParts, fmt.Sprintf("%s > $%d", p.Key, paramIdx))
			case "gte":
				whereParts = append(whereParts, fmt.Sprintf("%s >= $%d", p.Key, paramIdx))
			case "lt":
				whereParts = append(whereParts, fmt.Sprintf("%s < $%d", p.Key, paramIdx))
			case "lte":
				whereParts = append(whereParts, fmt.Sprintf("%s <= $%d", p.Key, paramIdx))
			case "in":
				whereParts = append(whereParts, fmt.Sprintf("%s = ANY($%d)", p.Key, paramIdx))
			case "neq":
				whereParts = append(whereParts, fmt.Sprintf("%s != $%d", p.Key, paramIdx))
			default:
				whereParts = append(whereParts, fmt.Sprintf("%s = $%d", p.Key, paramIdx))
			}
		}
	}

	dateStart, dateEnd := formatDateRange(getParam(q.QueryParam, "from"), getParam(q.QueryParam, "to"))
	if dateStart != nil && dateEnd == nil {
		args = append(args, *dateStart)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d", len(args)))
	} else if dateStart != nil && dateEnd != nil {
		args = append(args, *dateStart, *dateEnd)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d AND created_at < $%d", len(args)-1, len(args)))
	}

	for k, v := range cleanParams {
		if len(v) > 0 {
			args = append(args, v[0])
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", k, len(args)))
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	selectClause := "*"
	if sel, ok := q.QueryParam["select"]; ok && len(sel) > 0 {
		selects := strings.Join(strings.Split(sel[0], ","), ", ")
		if len(q.Fields) > 0 {
			selectClause = strings.Join(q.Fields, ", ") + ", " + selects
		} else {
			selectClause = selects
		}
	} else if len(q.Fields) > 0 {
		selectClause = strings.Join(q.Fields, ", ")
	}

	order := "DESC"
	if o, ok := q.QueryParam["order"]; ok && len(o) > 0 && o[0] == "asc" {
		order = "ASC"
	}
	sortField := "created_at"
	if s, ok := q.QueryParam["sort"]; ok && len(s) > 0 {
		sortField = s[0]
	}
	orderByClause := fmt.Sprintf("ORDER BY %s %s", sortField, order)

	page := 1
	limit := 50
	if p, ok := q.QueryParam["page"]; ok && len(p) > 0 {
		if v, err := strconv.Atoi(p[0]); err == nil && v > 0 {
			page = v
		}
	}
	if l, ok := q.QueryParam["limit"]; ok && len(l) > 0 {
		if v, err := strconv.Atoi(l[0]); err == nil && v > 0 {
			limit = v
		}
	}
	startIndex := (page - 1) * limit
	endIndex := page * limit

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", q.Table, whereClause)
	var total int64
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s %s LIMIT %d OFFSET %d",
		selectClause, q.Table, whereClause, orderByClause, limit, startIndex,
	)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns := q.Columns
	if len(columns) == 0 {
		columns = make([]string, len(rows.FieldDescriptions()))
		for i, fd := range rows.FieldDescriptions() {
			columns[i] = fd.Name
		}
	}

	var data []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	pagination := &PaginationResult{}
	if q.Paginate == "" || q.Paginate == "absolute" {
		if endIndex < int(total) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	} else if q.Paginate == "relative" {
		if endIndex < len(data) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	}

	return &PaginatedResult{
		Total:      total,
		Count:      len(data),
		Pagination: pagination,
		Data:       data,
	}, nil
}

func Populate(ctx context.Context, pool *pgxpool.Pool, q PopulateQuery) (*PaginatedResult, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	removeKeys := []string{"select", "sort", "page", "limit", "order", "from", "to"}
	cleanParams := make(map[string][]string)
	for k, v := range q.QueryParam {
		shouldRemove := false
		for _, rk := range removeKeys {
			if k == rk {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			cleanParams[k] = v
		}
	}

	var args []any
	var whereParts []string

	if q.Ref != "" && q.Value != nil {
		args = append(args, q.Value)
		refCol := q.Ref
		if refCol == "id" {
			refCol = "id"
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", refCol, len(args)))
	}

	dateStart, dateEnd := formatDateRange(getParam(q.QueryParam, "from"), getParam(q.QueryParam, "to"))
	if dateStart != nil && dateEnd == nil {
		args = append(args, *dateStart)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d", len(args)))
	} else if dateStart != nil && dateEnd != nil {
		args = append(args, *dateStart, *dateEnd)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d AND created_at < $%d", len(args)-1, len(args)))
	}

	for k, v := range cleanParams {
		if len(v) > 0 {
			args = append(args, v[0])
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", k, len(args)))
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	selectClause := "*"
	if sel, ok := q.QueryParam["select"]; ok && len(sel) > 0 {
		selects := strings.Join(strings.Split(sel[0], ","), ", ")
		if len(q.Fields) > 0 {
			selectClause = strings.Join(q.Fields, ", ") + ", " + selects
		} else {
			selectClause = selects
		}
	} else if len(q.Fields) > 0 {
		selectClause = strings.Join(q.Fields, ", ")
	}

	order := "DESC"
	if o, ok := q.QueryParam["order"]; ok && len(o) > 0 && o[0] == "asc" {
		order = "ASC"
	}
	sortField := "created_at"
	if s, ok := q.QueryParam["sort"]; ok && len(s) > 0 {
		sortField = s[0]
	}
	orderByClause := fmt.Sprintf("ORDER BY %s %s", sortField, order)

	page := 1
	limit := 50
	if p, ok := q.QueryParam["page"]; ok && len(p) > 0 {
		if v, err := strconv.Atoi(p[0]); err == nil && v > 0 {
			page = v
		}
	}
	if l, ok := q.QueryParam["limit"]; ok && len(l) > 0 {
		if v, err := strconv.Atoi(l[0]); err == nil && v > 0 {
			limit = v
		}
	}
	startIndex := (page - 1) * limit
	endIndex := page * limit

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", q.Table, whereClause)
	var total int64
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	joinClause := ""
	if q.Path != "" {
		joinClause = fmt.Sprintf("LEFT JOIN %s ON %s.id = %s.%s", q.Path, q.Table, q.Path, q.Ref)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s %s %s LIMIT %d OFFSET %d",
		selectClause, q.Table, joinClause, whereClause, orderByClause, limit, startIndex,
	)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns := q.Columns
	if len(columns) == 0 {
		columns = make([]string, len(rows.FieldDescriptions()))
		for i, fd := range rows.FieldDescriptions() {
			columns[i] = fd.Name
		}
	}

	var data []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	pagination := &PaginationResult{}
	if q.Paginate == "" || q.Paginate == "absolute" {
		if endIndex < int(total) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	} else if q.Paginate == "relative" {
		if endIndex < len(data) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	}

	return &PaginatedResult{
		Total:      total,
		Count:      len(data),
		Pagination: pagination,
		Data:       data,
	}, nil
}

func Search(ctx context.Context, pool *pgxpool.Pool, q SearchQuery) (*PaginatedResult, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	removeKeys := []string{"select", "sort", "page", "limit", "order", "from", "to"}
	cleanParams := make(map[string][]string)
	for k, v := range q.QueryParam {
		shouldRemove := false
		for _, rk := range removeKeys {
			if k == rk {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			cleanParams[k] = v
		}
	}

	var args []any
	var whereParts []string

	if q.Ref != "" && q.Value != nil {
		args = append(args, q.Value)
		refCol := q.Ref
		if refCol == "id" {
			refCol = "id"
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", refCol, len(args)))
	}

	dateStart, dateEnd := formatDateRange(getParam(q.QueryParam, "from"), getParam(q.QueryParam, "to"))
	if dateStart != nil && dateEnd == nil {
		args = append(args, *dateStart)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d", len(args)))
	} else if dateStart != nil && dateEnd != nil {
		args = append(args, *dateStart, *dateEnd)
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d AND created_at < $%d", len(args)-1, len(args)))
	}

	if len(q.Data) > 0 {
		var dataParts []string
		for _, p := range q.Data {
			args = append(args, p.Value)
			paramIdx := len(args)
			switch p.Operator {
			case "gt":
				dataParts = append(dataParts, fmt.Sprintf("%s > $%d", p.Key, paramIdx))
			case "gte":
				dataParts = append(dataParts, fmt.Sprintf("%s >= $%d", p.Key, paramIdx))
			case "lt":
				dataParts = append(dataParts, fmt.Sprintf("%s < $%d", p.Key, paramIdx))
			case "lte":
				dataParts = append(dataParts, fmt.Sprintf("%s <= $%d", p.Key, paramIdx))
			case "in":
				dataParts = append(dataParts, fmt.Sprintf("%s = ANY($%d)", p.Key, paramIdx))
			case "neq":
				dataParts = append(dataParts, fmt.Sprintf("%s != $%d", p.Key, paramIdx))
			default:
				dataParts = append(dataParts, fmt.Sprintf("%s = $%d", p.Key, paramIdx))
			}
		}
		dataWhere := strings.Join(dataParts, " AND ")

		switch q.Operator {
		case "or":
			whereParts = append(whereParts, "("+strings.ReplaceAll(dataWhere, " AND ", " OR ")+")")
		case "andor":
			var orParts []string
			if len(q.Query) > 0 {
				var qParts []string
				for _, p := range q.Query {
					args = append(args, p.Value)
					qParts = append(qParts, fmt.Sprintf("%s = $%d", p.Key, len(args)))
				}
				orParts = append(orParts, "("+strings.Join(qParts, " AND ")+")")
			}
			if dataWhere != "" {
				orParts = append(orParts, "("+dataWhere+")")
			}
			if len(orParts) > 0 {
				whereParts = append(whereParts, "("+strings.Join(orParts, " OR ")+")")
			}
		case "orand":
			var andParts []string
			if len(q.Query) > 0 {
				var qParts []string
				for _, p := range q.Query {
					args = append(args, p.Value)
					qParts = append(qParts, fmt.Sprintf("%s = $%d", p.Key, len(args)))
				}
				andParts = append(andParts, "("+strings.Join(qParts, " AND ")+")")
			}
			if dataWhere != "" {
				andParts = append(andParts, "("+dataWhere+")")
			}
			if len(andParts) > 0 {
				whereParts = append(whereParts, "("+strings.Join(andParts, " AND ")+")")
			}
		default:
			if dataWhere != "" {
				whereParts = append(whereParts, dataWhere)
			}
		}
	}

	for k, v := range cleanParams {
		if len(v) > 0 {
			args = append(args, v[0])
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", k, len(args)))
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	selectClause := "*"
	if sel, ok := q.QueryParam["select"]; ok && len(sel) > 0 {
		selects := strings.Join(strings.Split(sel[0], ","), ", ")
		if len(q.Fields) > 0 {
			selectClause = strings.Join(q.Fields, ", ") + ", " + selects
		} else {
			selectClause = selects
		}
	} else if len(q.Fields) > 0 {
		selectClause = strings.Join(q.Fields, ", ")
	}

	order := "DESC"
	if o, ok := q.QueryParam["order"]; ok && len(o) > 0 && o[0] == "asc" {
		order = "ASC"
	}
	sortField := "created_at"
	if s, ok := q.QueryParam["sort"]; ok && len(s) > 0 {
		sortField = s[0]
	}
	orderByClause := fmt.Sprintf("ORDER BY %s %s", sortField, order)

	page := 1
	limit := 50
	if p, ok := q.QueryParam["page"]; ok && len(p) > 0 {
		if v, err := strconv.Atoi(p[0]); err == nil && v > 0 {
			page = v
		}
	}
	if l, ok := q.QueryParam["limit"]; ok && len(l) > 0 {
		if v, err := strconv.Atoi(l[0]); err == nil && v > 0 {
			limit = v
		}
	}
	startIndex := (page - 1) * limit
	endIndex := page * limit

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", q.Table, whereClause)
	var total int64
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s %s LIMIT %d OFFSET %d",
		selectClause, q.Table, whereClause, orderByClause, limit, startIndex,
	)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns := q.Columns
	if len(columns) == 0 {
		columns = make([]string, len(rows.FieldDescriptions()))
		for i, fd := range rows.FieldDescriptions() {
			columns[i] = fd.Name
		}
	}

	var data []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	pagination := &PaginationResult{}
	if q.Paginate == "" || q.Paginate == "absolute" {
		if endIndex < int(total) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	} else if q.Paginate == "relative" {
		if endIndex < len(data) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	}

	return &PaginatedResult{
		Total:      total,
		Count:      len(data),
		Pagination: pagination,
		Data:       data,
	}, nil
}

func GeoSearch(ctx context.Context, pool *pgxpool.Pool, q GeoSearchQuery) (*PaginatedResult, error) {
	if pool == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	removeKeys := []string{"select", "sort", "page", "limit", "order", "from", "to"}
	cleanParams := make(map[string][]string)
	for k, v := range q.QueryParam {
		shouldRemove := false
		for _, rk := range removeKeys {
			if k == rk {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			cleanParams[k] = v
		}
	}

	var args []any
	var whereParts []string

	if q.Ref != "" && q.Value != nil {
		args = append(args, q.Value)
		refCol := q.Ref
		if refCol == "id" {
			refCol = "id"
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", refCol, len(args)))
	}

	switch q.Operator {
	case "center":
		args = append(args, q.GeoData.Longitude, q.GeoData.Latitude, q.GeoData.Radius)
		whereParts = append(whereParts, fmt.Sprintf(
			"ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			len(args)-2, len(args)-1, len(args),
		))
	case "centersphere":
		calcRadius := q.GeoData.Radius
		if q.GeoData.RadiusUnit == "km" {
			calcRadius = q.GeoData.Radius / earthRadiusKM
		} else {
			calcRadius = q.GeoData.Radius / earthRadiusML
		}
		args = append(args, q.GeoData.Longitude, q.GeoData.Latitude, calcRadius)
		whereParts = append(whereParts, fmt.Sprintf(
			"ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			len(args)-2, len(args)-1, len(args),
		))
	case "range":
		args = append(args, q.GeoData.Longitude, q.GeoData.Latitude, q.GeoData.MinDistance, q.GeoData.MaxDistance)
		whereParts = append(whereParts, fmt.Sprintf(
			"ST_Distance(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography) BETWEEN $%d AND $%d",
			len(args)-3, len(args)-2, len(args)-1, len(args),
		))
	case "within-poly":
		if len(q.GeoData.Coordinates) > 0 {
			coords := make([]string, len(q.GeoData.Coordinates))
			for i, c := range q.GeoData.Coordinates {
				coords[i] = fmt.Sprintf("%f %f", c[0], c[1])
			}
			wkt := "POLYGON((" + strings.Join(coords, ",") + "))"
			args = append(args, wkt)
			whereParts = append(whereParts, fmt.Sprintf(
				"ST_Within(location::geography, ST_GeomFromText($%d, 4326)::geography)",
				len(args),
			))
		}
	}

	if len(q.Data) > 0 {
		for _, p := range q.Data {
			args = append(args, p.Value)
			paramIdx := len(args)
			switch p.Operator {
			case "gt":
				whereParts = append(whereParts, fmt.Sprintf("%s > $%d", p.Key, paramIdx))
			case "gte":
				whereParts = append(whereParts, fmt.Sprintf("%s >= $%d", p.Key, paramIdx))
			case "lt":
				whereParts = append(whereParts, fmt.Sprintf("%s < $%d", p.Key, paramIdx))
			case "lte":
				whereParts = append(whereParts, fmt.Sprintf("%s <= $%d", p.Key, paramIdx))
			case "in":
				whereParts = append(whereParts, fmt.Sprintf("%s = ANY($%d)", p.Key, paramIdx))
			case "neq":
				whereParts = append(whereParts, fmt.Sprintf("%s != $%d", p.Key, paramIdx))
			default:
				whereParts = append(whereParts, fmt.Sprintf("%s = $%d", p.Key, paramIdx))
			}
		}
	}

	for k, v := range cleanParams {
		if len(v) > 0 {
			args = append(args, v[0])
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", k, len(args)))
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	selectClause := "*"
	if sel, ok := q.QueryParam["select"]; ok && len(sel) > 0 {
		selectClause = strings.Join(strings.Split(sel[0], ","), ", ")
	}

	order := "DESC"
	if o, ok := q.QueryParam["order"]; ok && len(o) > 0 && o[0] == "asc" {
		order = "ASC"
	}
	sortField := "created_at"
	if s, ok := q.QueryParam["sort"]; ok && len(s) > 0 {
		sortField = s[0]
	}
	orderByClause := fmt.Sprintf("ORDER BY %s %s", sortField, order)

	page := 1
	limit := 50
	if p, ok := q.QueryParam["page"]; ok && len(p) > 0 {
		if v, err := strconv.Atoi(p[0]); err == nil && v > 0 {
			page = v
		}
	}
	if l, ok := q.QueryParam["limit"]; ok && len(l) > 0 {
		if v, err := strconv.Atoi(l[0]); err == nil && v > 0 {
			limit = v
		}
	}
	startIndex := (page - 1) * limit
	endIndex := page * limit

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", q.Table, whereClause)
	var total int64
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s %s LIMIT %d OFFSET %d",
		selectClause, q.Table, whereClause, orderByClause, limit, startIndex,
	)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns := q.Columns
	if len(columns) == 0 {
		columns = make([]string, len(rows.FieldDescriptions()))
		for i, fd := range rows.FieldDescriptions() {
			columns[i] = fd.Name
		}
	}

	var data []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	pagination := &PaginationResult{}
	if q.Paginate == "" || q.Paginate == "absolute" {
		if endIndex < int(total) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	} else if q.Paginate == "relative" {
		if endIndex < len(data) {
			pagination.Next = &PageRef{Page: page + 1, Limit: limit}
		}
		if startIndex > 0 {
			pagination.Prev = &PageRef{Page: page - 1, Limit: limit}
		}
	}

	return &PaginatedResult{
		Total:      total,
		Count:      len(data),
		Pagination: pagination,
		Data:       data,
	}, nil
}

func formatDateRange(from, to string) (start *time.Time, end *time.Time) {

	if from = strings.TrimSpace(from); from != "" && commonutils.CheckDateFormat(from) {
		value := commonutils.DateToday(from).Dayjs
		start = &value
	}

	if to = strings.TrimSpace(to); to != "" && commonutils.CheckDateFormat(to) {
		value := commonutils.DateToday(to).Dayjs.Add(24 * time.Hour)
		end = &value
	}

	return
}

func getParam(queryParam map[string][]string, key string) string {

	if v, ok := queryParam[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}
