package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/events"
	"github.com/labstack/echo/v4"
)

type DataHandler struct {
	Service    *services.DatasetService
	Governance *services.GovernanceService
	Publisher  events.Publisher
}

func NewDataHandler(service *services.DatasetService, governance *services.GovernanceService, publisher events.Publisher) *DataHandler {
	return &DataHandler{Service: service, Governance: governance, Publisher: publisher}
}

func (h *DataHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.Dataset
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.Create(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.created", map[string]any{"hackathon_id": hackathonID, "dataset_id": created.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *DataHandler) Get(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	ds, err := h.Service.GetByHackathon(c.Request().Context(), hackathonID)
	if err != nil {
		return handleServiceError(err)
	}
	if ds == nil {
		return echo.NewHTTPError(http.StatusNotFound, "dataset not found")
	}
	return c.JSON(http.StatusOK, ds)
}

func (h *DataHandler) Update(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input services.DatasetUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.updated", map[string]any{"hackathon_id": hackathonID, "dataset_id": updated.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *DataHandler) Delete(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	if err := h.Service.Delete(c.Request().Context(), hackathonID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "dataset not found to delete")
		}
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.deleted", map[string]any{"hackathon_id": hackathonID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.deleted", map[string]string{"hackathon_id": hackathonID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *DataHandler) CreateFile(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.DatasetFile
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.CreateFile(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.file.created", map[string]any{"hackathon_id": hackathonID, "file_id": created.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.file.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *DataHandler) ListFiles(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.ListFiles(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *DataHandler) GetFile(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	fileID, err := parseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetFile(c.Request().Context(), hackathonID, fileID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "data file not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *DataHandler) UpdateFile(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	fileID, err := parseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}
	var input services.DatasetFileUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.UpdateFile(c.Request().Context(), hackathonID, fileID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.file.updated", map[string]any{"hackathon_id": hackathonID, "file_id": updated.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.file.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *DataHandler) DeleteFile(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	fileID, err := parseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}
	if err := h.Service.DeleteFile(c.Request().Context(), hackathonID, fileID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "data file not found to delete")
		}
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.file.deleted", map[string]any{"hackathon_id": hackathonID, "file_id": fileID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.file.deleted", map[string]string{"id": fileID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *DataHandler) CreateVariable(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.DatasetVariable
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.CreateVariable(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.variable.created", map[string]any{"hackathon_id": hackathonID, "variable_id": created.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.variable.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *DataHandler) ListVariables(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.ListVariables(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *DataHandler) GetVariable(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	variableID, err := parseUUIDParam(c, "variableId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetVariable(c.Request().Context(), hackathonID, variableID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "variable not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *DataHandler) UpdateVariable(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	variableID, err := parseUUIDParam(c, "variableId")
	if err != nil {
		return err
	}
	var input services.DatasetVariableUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.UpdateVariable(c.Request().Context(), hackathonID, variableID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.variable.updated", map[string]any{"hackathon_id": hackathonID, "variable_id": updated.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.variable.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *DataHandler) DeleteVariable(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	variableID, err := parseUUIDParam(c, "variableId")
	if err != nil {
		return err
	}
	if err := h.Service.DeleteVariable(c.Request().Context(), hackathonID, variableID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "variable not found to delete")
		}
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.data.variable.deleted", map[string]any{"hackathon_id": hackathonID, "variable_id": variableID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.data.variable.deleted", map[string]string{"id": variableID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *DataHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *DataHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
	if h.Governance == nil {
		return
	}
	raw, _ := json.Marshal(payload)
	_ = h.Governance.AppendAudit(c.Request().Context(), models.AuditLog{
		HackathonID: hackathonID,
		ActorID:     actorID,
		Action:      action,
		Payload:     raw,
	})
}
