package datasource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"google.golang.org/api/sheets/v4"
)

type spreadsheetService interface {
	Get(ctx context.Context, spreadSheetId string, sheetId string) ([][]interface{}, error)
}

type googleSpreadsheetService struct {
	spreadsheetService *sheets.SpreadsheetsService
}

func (s *googleSpreadsheetService) Get(ctx context.Context, spreadSheetId string, sheetName string) ([][]interface{}, error) {
	valueRange, err := s.spreadsheetService.Values.Get(spreadSheetId, sheetName).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return valueRange.Values, nil
}

func newGoogleSpreadsheetService(c *http.Client) (spreadsheetService, error) {
	service, err := sheets.New(c)
	if err != nil {
		return nil, err
	}
	return &googleSpreadsheetService{
		spreadsheetService: service.Spreadsheets,
	}, nil
}

type SpreadsheetDatasource struct {
	SpreadSheetID  string `json:"spreadsheet_id"`
	ColumnRowIndex int    `json:"column_row_index"`
	service        spreadsheetService
}

// NewSpreadsheetDatasource is return SpreadsheetDatasource instance
func NewSpreadsheetDatasource(c *http.Client, spreadsheetId string, columnRowIndex int) (*SpreadsheetDatasource, error) {
	if columnRowIndex < 0 {
		return nil, fmt.Errorf("columnRowIndex is invalid value: %d", columnRowIndex)
	}
	spreadSheetService, err := newGoogleSpreadsheetService(c)
	if err != nil {
		return nil, err
	}
	return &SpreadsheetDatasource{
		SpreadSheetID:  spreadsheetId,
		ColumnRowIndex: columnRowIndex,
		service:        spreadSheetService,
	}, nil
}

// GetSchema is getting schema from spreadsheet
func (ds *SpreadsheetDatasource) GetSchema(ctx context.Context, name string) (*Schema, error) {
	values, err := ds.getValues(ctx, name)
	if err != nil {
		return nil, err
	}
	primaryKey := &Key{
		KeyType: KeyTypePrimary,
	}
	cols := make([]*Column, 0)
	for rowIndex, row := range values {
		if rowIndex != ds.ColumnRowIndex {
			continue
		}
		for colIndex := range row {
			colName := row[colIndex].(string)
			// check primarykey
			reg := regexp.MustCompile("\\((.+?)\\)")
			if res := reg.FindStringSubmatch(colName); len(res) >= 2 {
				colName = res[1]
				primaryKey.ColumnNames = append(primaryKey.ColumnNames, colName)
			}
			cols = append(cols, &Column{
				Name:            colName,
				OrdinalPosition: colIndex,
				Type:            ColumnTypeString,
			})
		}
		break
	}
	return &Schema{
		Name:       name,
		PrimaryKey: primaryKey,
		Columns:    cols,
	}, nil
}

// GetRows is getting rows from spreadsheet
func (ds *SpreadsheetDatasource) GetRows(ctx context.Context, schema *Schema) ([]*Row, error) {
	values, err := ds.getValues(ctx, schema.Name)
	if err != nil {
		return nil, err
	}
	if len(values) > ds.ColumnRowIndex {
		valuesWithoutColumn := make([][]interface{}, len(values)-1)
		for rowIndex, row := range values {
			if rowIndex < ds.ColumnRowIndex {
				valuesWithoutColumn[rowIndex] = row
			} else if rowIndex > ds.ColumnRowIndex {
				valuesWithoutColumn[rowIndex-1] = row
			}
		}
		values = valuesWithoutColumn
	}
	rows := make([]*Row, len(values))
	for rowIndex, row := range values {
		rowValues := make(RowValues, len(schema.Columns))
		groupByKey := make(GroupByKey)
		for colIndex, col := range schema.Columns {
			var colValue *GenericColumnValue
			if colIndex < len(row) {
				colValue = NewStringGenericColumnValue(col, row[colIndex].(string))
			} else {
				colValue = NewStringGenericColumnValue(col, "")
			}
			rowValues[col.Name] = colValue
			// grouping primarykey
			for i := range schema.PrimaryKey.ColumnNames {
				if schema.PrimaryKey.ColumnNames[i] == col.Name {
					key := schema.PrimaryKey.String()
					groupByKey[key] = append(groupByKey[key], colValue)
				}
			}
		}
		rows[rowIndex] = &Row{GroupByKey: groupByKey, Values: rowValues}
	}
	return rows, nil
}

// SetSchema is set schema method
func (ds *SpreadsheetDatasource) SetSchema(ctx context.Context, schema *Schema) error {
	return errors.New("feature support")
}

// SetRows is set rows method
func (ds *SpreadsheetDatasource) SetRows(ctx context.Context, schema *Schema, rows []*Row) error {
	return errors.New("feature support")
}

func (ds *SpreadsheetDatasource) getValues(ctx context.Context, sheetName string) ([][]interface{}, error) {
	return ds.service.Get(ctx, ds.SpreadSheetID, sheetName)
}
