// SPDX-License-Identifier: GPL-3.0-or-later
package persistence

import (
	"context"
	"strings"

	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog/log"
)

func (db *DB) QueryJobs(ctx context.Context, apiQ api.JobsQuery) ([]*Job, error) {
	logger := log.Ctx(ctx)

	logger.Debug().Interface("q", apiQ).Msg("querying jobs")

	q := db.gormDB.WithContext(ctx).Model(&Job{})

	// WHERE
	if apiQ.StatusIn != nil {
		q = q.Where("status in ?", *apiQ.StatusIn)
	}
	if apiQ.Settings != nil {
		for setting, value := range apiQ.Settings.AdditionalProperties {
			q = q.Where("json_extract(metadata, ?) = ?", "$."+setting, value)
		}
	}
	if apiQ.Metadata != nil {
		for setting, value := range apiQ.Metadata.AdditionalProperties {
			if strings.ContainsRune(value, '%') {
				q = q.Where("json_extract(metadata, ?) like ?", "$."+setting, value)
			} else {
				q = q.Where("json_extract(metadata, ?) = ?", "$."+setting, value)
			}
		}
	}

	// OFFSET
	if apiQ.Offset != nil {
		q = q.Offset(*apiQ.Offset)
	}

	// LIMIT
	if apiQ.Limit != nil {
		q = q.Limit(*apiQ.Limit)
	}

	// ORDER BY
	if apiQ.OrderBy != nil {
		sqlOrder := ""
		for _, order := range *apiQ.OrderBy {
			if order == "" {
				continue
			}
			switch order[0] {
			case '-':
				sqlOrder = order[1:] + " desc"
			case '+':
				sqlOrder = order[1:] + " asc"
			default:
				sqlOrder = order
			}
			q = q.Order(sqlOrder)
		}
	}

	result := []*Job{}
	tx := q.Scan(&result)
	return result, tx.Error
}
