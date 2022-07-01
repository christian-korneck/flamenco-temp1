package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLastRendered(t *testing.T) {
	ctx, close, db, job1, _ := jobTasksTestFixtures(t)
	defer close()

	authoredJob2 := authorTestJob("1295757b-e668-4c49-8b89-f73db8270e42", "just-a-job")
	job2 := persistAuthoredJob(t, ctx, db, authoredJob2)

	assert.NoError(t, db.SetLastRendered(ctx, job1))
	{
		entries := []LastRendered{}
		db.gormDB.Model(&LastRendered{}).Scan(&entries)
		if assert.Len(t, entries, 1) {
			assert.Equal(t, job1.ID, entries[0].JobID, "job 1 should be the last-rendered one")
		}
	}

	assert.NoError(t, db.SetLastRendered(ctx, job2))
	{
		entries := []LastRendered{}
		db.gormDB.Model(&LastRendered{}).Scan(&entries)
		if assert.Len(t, entries, 1) {
			assert.Equal(t, job2.ID, entries[0].JobID, "job 2 should be the last-rendered one")
		}
	}
}

func TestGetLastRenderedJobUUID(t *testing.T) {
	ctx, close, db, job1, _ := jobTasksTestFixtures(t)
	defer close()

	{
		// Test without any renders.
		lastUUID, err := db.GetLastRenderedJobUUID(ctx)
		if assert.NoError(t, err, "absence of renders should not cause an error") {
			assert.Empty(t, lastUUID)
		}
	}

	{
		// Test with first render.
		assert.NoError(t, db.SetLastRendered(ctx, job1))
		lastUUID, err := db.GetLastRenderedJobUUID(ctx)
		if assert.NoError(t, err) {
			assert.Equal(t, job1.UUID, lastUUID)
		}
	}

	{
		// Test with 2nd or subsequent render.
		authoredJob2 := authorTestJob("1295757b-e668-4c49-8b89-f73db8270e42", "just-a-job")
		job2 := persistAuthoredJob(t, ctx, db, authoredJob2)

		assert.NoError(t, db.SetLastRendered(ctx, job2))
		lastUUID, err := db.GetLastRenderedJobUUID(ctx)
		if assert.NoError(t, err) {
			assert.Equal(t, job2.UUID, lastUUID)
		}
	}
}
