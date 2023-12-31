package sqlstmt

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestColumnMapper_Columns(t *testing.T) {
	td := testingDataType{}
	require.NotEmpty(t, td.ColumnMapperMap().Columns())
}

func TestMapColumn(t *testing.T) {
	db := testConnectToDatabase(t)
	defer testCloseDB(t, db)

	stmt, err := ConvertNamedToPositionalParams(
		insertTestingDataTypeQuery,
	)
	require.NoError(t, err)
	require.NotNil(t, stmt)

	testData := genFakeTestingDataType(t)

	result, execErr := stmt.Exec(
		db,
		testData.asParameters()...,
	)
	require.NoError(t, execErr)
	require.NotNil(t, result)

	t.Run("valid types", func(t *testing.T) {
		stmt, err = ConvertNamedToPositionalParams(selectTestingDataTypeQuery)
		require.NoError(t, err)
		require.NotNil(t, stmt)

		row, err := stmt.QueryRow(db,
			SetParameter("uuids", pq.Array([]string{testData.UUID})),
		)
		require.NoError(t, err)
		require.NotNil(t, row)
		columns := []string{
			"testing_datatype_id",
			"testing_datatype_uuid",
			"word",
			"paragraph",
			"metadata",
			"created_at",
		}
		mappedRow, err := MapRow(row, columns)
		require.NoError(t, err)
		require.NotNil(t, mappedRow)

		td := testingDataType{}
		err = td.mapRow(mappedRow)
		require.NoError(t, err)

		require.NotEmpty(t, td.ID)
		require.Equal(t, testData.UUID, td.UUID)
		require.Equal(t, testData.Word, td.Word)
		require.Equal(t, testData.Paragraph, td.Paragraph)
	})

	t.Run("invalid type", func(t *testing.T) {
		invalidDataType := []byte(`
			SELECT
				CAST(td.testing_datatype_id AS VARCHAR) AS test_data_type_id,
				td.testing_datatype_uuid,
				td.word,
				td.paragraph,
				td.metadata,
				td.created_at
			FROM testing_datatypes td
			WHERE (nullif(:id, NULL) IS NULL OR td.testing_datatype_id = :id)
			AND (nullif(:ids, '{}') IS NULL OR td.testing_datatype_id = ANY(:ids))
			AND (nullif(:uuid, NULL) IS NULL OR td.testing_datatype_uuid = :uuid)
			AND (nullif(:uuids, '{}') IS NULL OR td.testing_datatype_uuid = ANY(:uuids))
			ORDER BY td.created_at
		`)

		stmt, err = ConvertNamedToPositionalParams(invalidDataType)
		require.NoError(t, err)
		require.NotNil(t, stmt)

		row, err := stmt.QueryRow(db,
			SetParameter("uuids", pq.Array([]string{testData.UUID})),
		)
		require.NoError(t, err)
		require.NotNil(t, row)
		columns := []string{
			"testing_datatype_id",
			"testing_datatype_uuid",
			"word",
			"paragraph",
			"metadata",
			"created_at",
		}
		mappedRow, err := MapRow(row, columns)
		require.NoError(t, err)
		require.NotNil(t, mappedRow)

		td := testingDataType{}
		err = td.mapRow(mappedRow)
		require.Error(t, err)
		require.Equal(t, err.Error(), "column testing_datatype_id has a type of string and does not match asserted type: int64")
	})
}
