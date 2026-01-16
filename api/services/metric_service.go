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

type MetricService struct {
	DB *sql.DB
}

func NewMetricService(db *sql.DB) *MetricService {
	return &MetricService{DB: db}
}

func (s *MetricService) Create(ctx context.Context, hackathonID string, input models.EvaluationMetric) (*models.EvaluationMetric, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	metric, err := buildMetric(hackathonID, input)
	if err != nil {
		return nil, err
	}
	if metric.Scope == models.MetricScopePerTarget {
		if err := s.ensureTargetVariable(ctx, hackathonID, metric.TargetVariable); err != nil {
			return nil, err
		}
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO evaluation_metrics (id, hackathon_id, name, metric_type, direction, scope, target_variable, weight, description, params, is_primary, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		metric.ID, metric.HackathonID, metric.Name, metric.MetricType, metric.Direction, metric.Scope, nullableString(metric.TargetVariable), metric.Weight, metric.Description, metric.Params, metric.IsPrimary, metric.CreatedAt, metric.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}

	if metric.IsPrimary {
		_ = s.clearPrimary(ctx, metric.HackathonID, metric.ID)
	}

	return &metric, nil
}

func (s *MetricService) List(ctx context.Context, hackathonID string, limit, offset int) ([]models.EvaluationMetric, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, name, metric_type, direction, scope, target_variable, weight, description, params, is_primary, created_at, updated_at
		FROM evaluation_metrics
		WHERE hackathon_id = $1
		ORDER BY created_at
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.EvaluationMetric
	for rows.Next() {
		var m models.EvaluationMetric
		var params []byte
		var target sql.NullString
		if err := rows.Scan(&m.ID, &m.HackathonID, &m.Name, &m.MetricType, &m.Direction, &m.Scope, &target, &m.Weight, &m.Description, &params, &m.IsPrimary, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		if target.Valid {
			m.TargetVariable = target.String
		}
		m.Params = params
		items = append(items, m)
	}
	return items, nil
}

func (s *MetricService) GetByID(ctx context.Context, hackathonID, metricID string) (*models.EvaluationMetric, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, name, metric_type, direction, scope, target_variable, weight, description, params, is_primary, created_at, updated_at
		FROM evaluation_metrics
		WHERE id = $1 AND hackathon_id = $2`, metricID, hackathonID)

	var m models.EvaluationMetric
	var params []byte
	var target sql.NullString
	if err := row.Scan(&m.ID, &m.HackathonID, &m.Name, &m.MetricType, &m.Direction, &m.Scope, &target, &m.Weight, &m.Description, &params, &m.IsPrimary, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	if target.Valid {
		m.TargetVariable = target.String
	}
	m.Params = params
	return &m, nil
}

type MetricUpdateInput struct {
	Name           *string          `json:"name,omitempty"`
	MetricType     *string          `json:"metric_type,omitempty"`
	Direction      *string          `json:"direction,omitempty"`
	Scope          *string          `json:"scope,omitempty"`
	TargetVariable *string          `json:"target_variable,omitempty"`
	Weight         *float64         `json:"weight,omitempty"`
	Description    *string          `json:"description,omitempty"`
	Params         *json.RawMessage `json:"params,omitempty"`
	IsPrimary      *bool            `json:"is_primary,omitempty"`
}

func (s *MetricService) Update(ctx context.Context, hackathonID, metricID string, input MetricUpdateInput) (*models.EvaluationMetric, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	existing, err := s.GetByID(ctx, hackathonID, metricID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("metric not found: %w", ErrNotFound)
	}

	name := existing.Name
	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			return nil, fmt.Errorf("metric name is required: %w", ErrInvalid)
		}
		name = strings.TrimSpace(*input.Name)
	}
	metricType := existing.MetricType
	if input.MetricType != nil {
		mt := normalizeMetricType(*input.MetricType)
		if mt == "" {
			return nil, fmt.Errorf("metric_type is required: %w", ErrInvalid)
		}
		if !isAllowedMetricType(mt) {
			return nil, fmt.Errorf("unsupported metric_type: %w", ErrInvalid)
		}
		metricType = mt
	}
	direction := existing.Direction
	if input.Direction != nil {
		dir := normalizeDirection(*input.Direction)
		if dir == "" {
			return nil, fmt.Errorf("direction is required: %w", ErrInvalid)
		}
		if !isAllowedDirection(dir) {
			return nil, fmt.Errorf("unsupported direction: %w", ErrInvalid)
		}
		direction = dir
	}
	scope := existing.Scope
	targetVariable := strings.TrimSpace(existing.TargetVariable)
	if input.Scope != nil {
		val := normalizeScope(*input.Scope)
		if val == "" {
			return nil, fmt.Errorf("scope is required: %w", ErrInvalid)
		}
		if !isAllowedScope(val) {
			return nil, fmt.Errorf("unsupported scope: %w", ErrInvalid)
		}
		scope = val
	}
	if input.TargetVariable != nil {
		targetVariable = strings.TrimSpace(*input.TargetVariable)
	}
	if input.Scope == nil && input.TargetVariable != nil && targetVariable != "" {
		scope = models.MetricScopePerTarget
	}
	if scope == "" {
		scope = models.MetricScopeOverall
	}
	if scope == models.MetricScopePerTarget {
		if targetVariable == "" {
			return nil, fmt.Errorf("target_variable is required for per_target metrics: %w", ErrInvalid)
		}
		if err := s.ensureTargetVariable(ctx, hackathonID, targetVariable); err != nil {
			return nil, err
		}
	} else if targetVariable != "" {
		return nil, fmt.Errorf("target_variable is only allowed for per_target metrics: %w", ErrInvalid)
	}
	weight := existing.Weight
	if input.Weight != nil {
		if *input.Weight < 0 {
			return nil, fmt.Errorf("weight must be >= 0: %w", ErrInvalid)
		}
		weight = *input.Weight
	}
	description := existing.Description
	if input.Description != nil {
		description = *input.Description
	}
	params := existing.Params
	if input.Params != nil {
		if err := ensureValidJSON(*input.Params, "params"); err != nil {
			return nil, err
		}
		params = normalizeMetadata(*input.Params)
	}
	isPrimary := existing.IsPrimary
	if input.IsPrimary != nil {
		isPrimary = *input.IsPrimary
	}
	if isPrimary && weight <= 0 {
		return nil, fmt.Errorf("primary metric weight must be > 0: %w", ErrInvalid)
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE evaluation_metrics
		SET name = $1, metric_type = $2, direction = $3, scope = $4, target_variable = $5, weight = $6, description = $7, params = $8, is_primary = $9, updated_at = NOW()
		WHERE id = $10 AND hackathon_id = $11`,
		name, metricType, direction, scope, nullableString(targetVariable), weight, description, params, isPrimary, metricID, hackathonID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	if isPrimary {
		_ = s.clearPrimary(ctx, hackathonID, metricID)
	}
	return s.GetByID(ctx, hackathonID, metricID)
}

func (s *MetricService) Delete(ctx context.Context, hackathonID, metricID string) error {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return err
	}
	res, err := s.DB.ExecContext(ctx, `DELETE FROM evaluation_metrics WHERE id = $1 AND hackathon_id = $2`, metricID, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("metric not found: %w", ErrNotFound)
	}
	return nil
}

func (s *MetricService) clearPrimary(ctx context.Context, hackathonID, keepID string) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE evaluation_metrics
		SET is_primary = false
		WHERE hackathon_id = $1 AND id <> $2`, hackathonID, keepID)
	return mapSQLError(err)
}

func buildMetric(hackathonID string, input models.EvaluationMetric) (models.EvaluationMetric, error) {
	name := strings.TrimSpace(input.Name)
	metricType := normalizeMetricType(input.MetricType)
	direction := normalizeDirection(input.Direction)
	scope := normalizeScope(input.Scope)
	targetVariable := strings.TrimSpace(input.TargetVariable)
	weight := input.Weight
	if weight == 0 {
		weight = 1
	}
	if name == "" || metricType == "" || direction == "" {
		return models.EvaluationMetric{}, fmt.Errorf("name, metric_type, and direction are required: %w", ErrInvalid)
	}
	if !isAllowedMetricType(metricType) {
		return models.EvaluationMetric{}, fmt.Errorf("unsupported metric_type: %w", ErrInvalid)
	}
	if !isAllowedDirection(direction) {
		return models.EvaluationMetric{}, fmt.Errorf("unsupported direction: %w", ErrInvalid)
	}
	if scope == "" && targetVariable != "" {
		scope = models.MetricScopePerTarget
	}
	if scope == "" {
		scope = models.MetricScopeOverall
	}
	if !isAllowedScope(scope) {
		return models.EvaluationMetric{}, fmt.Errorf("unsupported scope: %w", ErrInvalid)
	}
	if scope == models.MetricScopePerTarget && targetVariable == "" {
		return models.EvaluationMetric{}, fmt.Errorf("target_variable is required for per_target metrics: %w", ErrInvalid)
	}
	if scope != models.MetricScopePerTarget && targetVariable != "" {
		return models.EvaluationMetric{}, fmt.Errorf("target_variable is only allowed for per_target metrics: %w", ErrInvalid)
	}
	if weight < 0 {
		return models.EvaluationMetric{}, fmt.Errorf("weight must be >= 0: %w", ErrInvalid)
	}
	if input.IsPrimary && weight <= 0 {
		return models.EvaluationMetric{}, fmt.Errorf("primary metric weight must be > 0: %w", ErrInvalid)
	}
	if err := ensureValidJSON(input.Params, "params"); err != nil {
		return models.EvaluationMetric{}, err
	}

	now := time.Now().UTC()
	metric := models.EvaluationMetric{
		ID:             uuid.NewString(),
		HackathonID:    hackathonID,
		Name:           name,
		MetricType:     metricType,
		Direction:      direction,
		Scope:          scope,
		TargetVariable: targetVariable,
		Weight:         weight,
		Description:    input.Description,
		Params:         normalizeMetadata(input.Params),
		IsPrimary:      input.IsPrimary,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return metric, nil
}

func normalizeMetricType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeDirection(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeScope(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isAllowedDirection(value string) bool {
	switch value {
	case models.MetricDirectionMinimize, models.MetricDirectionMaximize:
		return true
	default:
		return false
	}
}

func isAllowedScope(value string) bool {
	switch value {
	case models.MetricScopeOverall, models.MetricScopePerTarget:
		return true
	default:
		return false
	}
}

var allowedMetricTypes = map[string]struct{}{
	"mae":               {},
	"mse":               {},
	"rmse":              {},
	"rmsle":             {},
	"mape":              {},
	"smape":             {},
	"r2":                {},
	"accuracy":          {},
	"balanced_accuracy": {},
	"precision":         {},
	"recall":            {},
	"f1":                {},
	"f_beta":            {},
	"roc_auc":           {},
	"pr_auc":            {},
	"log_loss":          {},
	"mcc":               {},
	"kappa":             {},
	"top_k_accuracy":    {},
	"map":               {},
	"map50":             {},
	"map50_95":          {},
	"ndcg":              {},
	"mrr":               {},
	"iou":               {},
	"dice":              {},
	"pixel_accuracy":    {},
	"bleu":              {},
	"rouge":             {},
	"wer":               {},
	"cer":               {},
	"perplexity":        {},
	"psnr":              {},
	"ssim":              {},
	"fid":               {},
	"lpips":             {},
	"clip_score":        {},
	"custom":            {},
}

func isAllowedMetricType(value string) bool {
	_, ok := allowedMetricTypes[value]
	return ok
}

func (s *MetricService) ensureTargetVariable(ctx context.Context, hackathonID, name string) error {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM dataset_variables dv
			JOIN hackathon_datasets hd ON dv.dataset_id = hd.id
			WHERE hd.hackathon_id = $1 AND dv.name = $2 AND dv.role = $3
		)`, hackathonID, name, models.DatasetVariableRoleTarget).Scan(&exists)
	if err != nil {
		return mapSQLError(err)
	}
	if !exists {
		return fmt.Errorf("target_variable not found in dataset: %w", ErrInvalid)
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
