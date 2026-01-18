package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/google/uuid"
)

type DatasetService struct {
	DB *sql.DB
}

func NewDatasetService(db *sql.DB) *DatasetService {
	return &DatasetService{DB: db}
}

func (s *DatasetService) Create(ctx context.Context, hackathonID string, input models.Dataset) (*models.Dataset, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	if input.Title == "" || input.Description == "" {
		return nil, fmt.Errorf("data title and description are required: %w", ErrInvalid)
	}
	sourceURLs, err := normalizeURLList(input.SourceURLs)
	if err != nil {
		return nil, err
	}
	if err := ensureValidJSON(input.ResponseSchema, "response_schema"); err != nil {
		return nil, err
	}
	sourceRaw, err := json.Marshal(sourceURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid source_urls: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	ds := models.Dataset{
		ID:             uuid.NewString(),
		HackathonID:    hackathonID,
		Title:          input.Title,
		Description:    input.Description,
		SourceURLs:     sourceURLs,
		ResponseSchema: normalizeMetadata(input.ResponseSchema),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO hackathon_datasets (id, hackathon_id, title, description, source_urls, response_schema, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		ds.ID, ds.HackathonID, ds.Title, ds.Description, sourceRaw, ds.ResponseSchema, ds.CreatedAt, ds.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &ds, nil
}

func (s *DatasetService) GetByHackathon(ctx context.Context, hackathonID string) (*models.Dataset, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, title, description, source_urls, response_schema, created_at, updated_at
		FROM hackathon_datasets
		WHERE hackathon_id = $1`, hackathonID)

	var ds models.Dataset
	var sourceRaw []byte
	var schema []byte
	if err := row.Scan(&ds.ID, &ds.HackathonID, &ds.Title, &ds.Description, &sourceRaw, &schema, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	if len(sourceRaw) > 0 {
		if err := json.Unmarshal(sourceRaw, &ds.SourceURLs); err != nil {
			return nil, mapSQLError(err)
		}
	}
	ds.ResponseSchema = schema
	return &ds, nil
}

type DatasetUpdateInput struct {
	Title          *string          `json:"title,omitempty"`
	Description    *string          `json:"description,omitempty"`
	SourceURLs     *[]string        `json:"source_urls,omitempty"`
	ResponseSchema *json.RawMessage `json:"response_schema,omitempty"`
}

func (s *DatasetService) Update(ctx context.Context, hackathonID string, input DatasetUpdateInput) (*models.Dataset, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	existing, err := s.GetByHackathon(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("dataset not found: %w", ErrNotFound)
	}

	title := existing.Title
	if input.Title != nil {
		if *input.Title == "" {
			return nil, fmt.Errorf("data title is required: %w", ErrInvalid)
		}
		title = *input.Title
	}
	description := existing.Description
	if input.Description != nil {
		if *input.Description == "" {
			return nil, fmt.Errorf("data description is required: %w", ErrInvalid)
		}
		description = *input.Description
	}
	sourceURLs := existing.SourceURLs
	if input.SourceURLs != nil {
		normalized, err := normalizeURLList(*input.SourceURLs)
		if err != nil {
			return nil, err
		}
		sourceURLs = normalized
	}
	schema := existing.ResponseSchema
	if input.ResponseSchema != nil {
		if err := ensureValidJSON(*input.ResponseSchema, "response_schema"); err != nil {
			return nil, err
		}
		schema = normalizeMetadata(*input.ResponseSchema)
	}
	sourceRaw, err := json.Marshal(sourceURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid source_urls: %w", ErrInvalid)
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE hackathon_datasets
		SET title = $1, description = $2, source_urls = $3, response_schema = $4, updated_at = NOW()
		WHERE hackathon_id = $5`,
		title, description, sourceRaw, schema, hackathonID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByHackathon(ctx, hackathonID)
}

func (s *DatasetService) Delete(ctx context.Context, hackathonID string) error {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return err
	}
	res, err := s.DB.ExecContext(ctx, `DELETE FROM hackathon_datasets WHERE hackathon_id = $1`, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("dataset not found: %w", ErrNotFound)
	}
	return nil
}

func (s *DatasetService) getDatasetID(ctx context.Context, hackathonID string) (string, error) {
	ds, err := s.GetByHackathon(ctx, hackathonID)
	if err != nil {
		return "", err
	}
	if ds == nil {
		return "", fmt.Errorf("dataset not found: %w", ErrNotFound)
	}
	return ds.ID, nil
}

func normalizeURLList(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			return nil, fmt.Errorf("source_urls entries must be non-empty: %w", ErrInvalid)
		}
		normalized = append(normalized, item)
	}
	return normalized, nil
}

func (s *DatasetService) CreateFile(ctx context.Context, hackathonID string, input models.DatasetFile) (*models.DatasetFile, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	fileType := normalizeFileType(input.FileType)
	if name == "" || fileType == "" || input.URL == "" {
		return nil, fmt.Errorf("file name, file_type, and url are required: %w", ErrInvalid)
	}
	if !isAllowedFileType(fileType) {
		return nil, fmt.Errorf("unsupported file_type: %w", ErrInvalid)
	}
	if input.SizeBytes < 0 {
		return nil, fmt.Errorf("size_bytes must be >= 0: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	file := models.DatasetFile{
		ID:          uuid.NewString(),
		DatasetID:   datasetID,
		Name:        name,
		FileType:    fileType,
		Description: input.Description,
		URL:         input.URL,
		SizeBytes:   input.SizeBytes,
		Checksum:    input.Checksum,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO dataset_files (id, dataset_id, name, file_type, description, url, size_bytes, checksum, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		file.ID, file.DatasetID, file.Name, file.FileType, file.Description, file.URL, file.SizeBytes, file.Checksum, file.CreatedAt, file.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &file, nil
}

func (s *DatasetService) ListFiles(ctx context.Context, hackathonID string, limit, offset int) ([]models.DatasetFile, error) {
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, dataset_id, name, file_type, description, url, size_bytes, checksum, created_at, updated_at
		FROM dataset_files
		WHERE dataset_id = $1
		ORDER BY created_at
		LIMIT $2 OFFSET $3`, datasetID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.DatasetFile
	for rows.Next() {
		var f models.DatasetFile
		if err := rows.Scan(&f.ID, &f.DatasetID, &f.Name, &f.FileType, &f.Description, &f.URL, &f.SizeBytes, &f.Checksum, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		items = append(items, f)
	}
	return items, nil
}

func (s *DatasetService) GetFile(ctx context.Context, hackathonID, fileID string) (*models.DatasetFile, error) {
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, dataset_id, name, file_type, description, url, size_bytes, checksum, created_at, updated_at
		FROM dataset_files
		WHERE id = $1 AND dataset_id = $2`, fileID, datasetID)

	var f models.DatasetFile
	if err := row.Scan(&f.ID, &f.DatasetID, &f.Name, &f.FileType, &f.Description, &f.URL, &f.SizeBytes, &f.Checksum, &f.CreatedAt, &f.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	return &f, nil
}

type DatasetFileUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	FileType    *string `json:"file_type,omitempty"`
	Description *string `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
	SizeBytes   *int64  `json:"size_bytes,omitempty"`
	Checksum    *string `json:"checksum,omitempty"`
}

func (s *DatasetService) UpdateFile(ctx context.Context, hackathonID, fileID string, input DatasetFileUpdateInput) (*models.DatasetFile, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	existing, err := s.GetFile(ctx, hackathonID, fileID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("data file not found: %w", ErrNotFound)
	}

	name := existing.Name
	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			return nil, fmt.Errorf("file name is required: %w", ErrInvalid)
		}
		name = strings.TrimSpace(*input.Name)
	}
	fileType := existing.FileType
	if input.FileType != nil {
		ft := normalizeFileType(*input.FileType)
		if ft == "" {
			return nil, fmt.Errorf("file_type is required: %w", ErrInvalid)
		}
		if !isAllowedFileType(ft) {
			return nil, fmt.Errorf("unsupported file_type: %w", ErrInvalid)
		}
		fileType = ft
	}
	description := existing.Description
	if input.Description != nil {
		description = *input.Description
	}
	url := existing.URL
	if input.URL != nil {
		if *input.URL == "" {
			return nil, fmt.Errorf("file url is required: %w", ErrInvalid)
		}
		url = *input.URL
	}
	sizeBytes := existing.SizeBytes
	if input.SizeBytes != nil {
		if *input.SizeBytes < 0 {
			return nil, fmt.Errorf("size_bytes must be >= 0: %w", ErrInvalid)
		}
		sizeBytes = *input.SizeBytes
	}
	checksum := existing.Checksum
	if input.Checksum != nil {
		checksum = *input.Checksum
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE dataset_files
		SET name = $1, file_type = $2, description = $3, url = $4, size_bytes = $5, checksum = $6, updated_at = NOW()
		WHERE id = $7 AND dataset_id = $8`,
		name, fileType, description, url, sizeBytes, checksum, fileID, existing.DatasetID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetFile(ctx, hackathonID, fileID)
}

func (s *DatasetService) DeleteFile(ctx context.Context, hackathonID, fileID string) error {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return err
	}
	existing, err := s.GetFile(ctx, hackathonID, fileID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("data file not found: %w", ErrNotFound)
	}
	res, err := s.DB.ExecContext(ctx, `DELETE FROM dataset_files WHERE id = $1 AND dataset_id = $2`, fileID, existing.DatasetID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("data file not found: %w", ErrNotFound)
	}
	return nil
}

func (s *DatasetService) CreateVariable(ctx context.Context, hackathonID string, input models.DatasetVariable) (*models.DatasetVariable, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	role := normalizeRole(input.Role)
	dataType := normalizeDataType(input.DataType)
	if name == "" || role == "" || dataType == "" {
		return nil, fmt.Errorf("variable name, role, and data_type are required: %w", ErrInvalid)
	}
	if !isAllowedRole(role) {
		return nil, fmt.Errorf("unsupported role: %w", ErrInvalid)
	}
	if !isAllowedDataType(dataType) {
		return nil, fmt.Errorf("unsupported data_type: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	variable := models.DatasetVariable{
		ID:          uuid.NewString(),
		DatasetID:   datasetID,
		Name:        name,
		Role:        role,
		DataType:    dataType,
		Description: input.Description,
		Unit:        input.Unit,
		Category:    input.Category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO dataset_variables (id, dataset_id, name, role, data_type, description, unit, category, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		variable.ID, variable.DatasetID, variable.Name, variable.Role, variable.DataType, variable.Description, variable.Unit, variable.Category, variable.CreatedAt, variable.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &variable, nil
}

func (s *DatasetService) ListVariables(ctx context.Context, hackathonID string, limit, offset int) ([]models.DatasetVariable, error) {
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, dataset_id, name, role, data_type, description, unit, category, created_at, updated_at
		FROM dataset_variables
		WHERE dataset_id = $1
		ORDER BY created_at
		LIMIT $2 OFFSET $3`, datasetID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.DatasetVariable
	for rows.Next() {
		var v models.DatasetVariable
		if err := rows.Scan(&v.ID, &v.DatasetID, &v.Name, &v.Role, &v.DataType, &v.Description, &v.Unit, &v.Category, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		items = append(items, v)
	}
	return items, nil
}

func (s *DatasetService) GetVariable(ctx context.Context, hackathonID, variableID string) (*models.DatasetVariable, error) {
	datasetID, err := s.getDatasetID(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, dataset_id, name, role, data_type, description, unit, category, created_at, updated_at
		FROM dataset_variables
		WHERE id = $1 AND dataset_id = $2`, variableID, datasetID)

	var v models.DatasetVariable
	if err := row.Scan(&v.ID, &v.DatasetID, &v.Name, &v.Role, &v.DataType, &v.Description, &v.Unit, &v.Category, &v.CreatedAt, &v.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	return &v, nil
}

type DatasetVariableUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Role        *string `json:"role,omitempty"`
	DataType    *string `json:"data_type,omitempty"`
	Description *string `json:"description,omitempty"`
	Unit        *string `json:"unit,omitempty"`
	Category    *string `json:"category,omitempty"`
}

func (s *DatasetService) UpdateVariable(ctx context.Context, hackathonID, variableID string, input DatasetVariableUpdateInput) (*models.DatasetVariable, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	existing, err := s.GetVariable(ctx, hackathonID, variableID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("variable not found: %w", ErrNotFound)
	}

	name := existing.Name
	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			return nil, fmt.Errorf("variable name is required: %w", ErrInvalid)
		}
		name = strings.TrimSpace(*input.Name)
	}
	role := existing.Role
	if input.Role != nil {
		val := normalizeRole(*input.Role)
		if val == "" {
			return nil, fmt.Errorf("role is required: %w", ErrInvalid)
		}
		if !isAllowedRole(val) {
			return nil, fmt.Errorf("unsupported role: %w", ErrInvalid)
		}
		role = val
	}
	dataType := existing.DataType
	if input.DataType != nil {
		val := normalizeDataType(*input.DataType)
		if val == "" {
			return nil, fmt.Errorf("data_type is required: %w", ErrInvalid)
		}
		if !isAllowedDataType(val) {
			return nil, fmt.Errorf("unsupported data_type: %w", ErrInvalid)
		}
		dataType = val
	}
	description := existing.Description
	if input.Description != nil {
		description = *input.Description
	}
	unit := existing.Unit
	if input.Unit != nil {
		unit = *input.Unit
	}
	category := existing.Category
	if input.Category != nil {
		category = *input.Category
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE dataset_variables
		SET name = $1, role = $2, data_type = $3, description = $4, unit = $5, category = $6, updated_at = NOW()
		WHERE id = $7 AND dataset_id = $8`,
		name, role, dataType, description, unit, category, variableID, existing.DatasetID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetVariable(ctx, hackathonID, variableID)
}

func (s *DatasetService) DeleteVariable(ctx context.Context, hackathonID, variableID string) error {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return err
	}
	existing, err := s.GetVariable(ctx, hackathonID, variableID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("variable not found: %w", ErrNotFound)
	}
	res, err := s.DB.ExecContext(ctx, `DELETE FROM dataset_variables WHERE id = $1 AND dataset_id = $2`, variableID, existing.DatasetID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("variable not found: %w", ErrNotFound)
	}
	return nil
}

func normalizeFileType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isAllowedFileType(value string) bool {
	switch value {
	case models.DatasetFileTypeTrain, models.DatasetFileTypeTest, models.DatasetFileTypeSampleSubmission,
		models.DatasetFileTypeDataDictionary, models.DatasetFileTypeOther:
		return true
	default:
		return false
	}
}

func normalizeRole(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isAllowedRole(value string) bool {
	switch value {
	case models.DatasetVariableRoleFeature, models.DatasetVariableRoleTarget, models.DatasetVariableRoleIdentifier,
		models.DatasetVariableRoleMetadata:
		return true
	default:
		return false
	}
}

func normalizeDataType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isAllowedDataType(value string) bool {
	switch value {
	case models.DatasetDataTypeString, models.DatasetDataTypeInteger, models.DatasetDataTypeFloat,
		models.DatasetDataTypeBoolean, models.DatasetDataTypeCategorical, models.DatasetDataTypeDatetime,
		models.DatasetDataTypeText:
		return true
	default:
		return false
	}
}
