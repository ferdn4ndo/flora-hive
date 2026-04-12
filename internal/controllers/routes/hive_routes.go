package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"flora-hive/internal/controllers/authctx"
	"flora-hive/internal/controllers/middleware"
	routemw "flora-hive/internal/controllers/routes/middlewares"
	"flora-hive/internal/domain/models"
	"flora-hive/internal/domain/mqtttopic"
	"flora-hive/internal/domain/ports"
	mqttsvc "flora-hive/internal/infrastructure/mqtt"
	"flora-hive/internal/infrastructure/userver"
	"flora-hive/internal/services"
	"flora-hive/lib"
)

// Module registers HTTP routes for Flora Hive.
var Module = fx.Options(
	routemw.Module,
	fx.Provide(NewHiveRoutes),
	fx.Provide(NewRoutes),
)

// Routes bundles route groups.
type Routes struct {
	routes  []Route
	handler lib.RequestHandler
}

// Route is a mountable route group.
type Route interface {
	Setup()
}

// NewRoutes builds the top-level router.
func NewRoutes(hive HiveRoutes, handler lib.RequestHandler) Routes {
	return Routes{
		routes:  []Route{hive},
		handler: handler,
	}
}

// Setup installs all routes.
func (r Routes) Setup() {
	for _, rt := range r.routes {
		rt.Setup()
	}
	r.handler.Gin.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "OK") })
	r.handler.Gin.GET("/favicon.ico", func(c *gin.Context) { c.String(http.StatusOK, "OK") })
	r.handler.Gin.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	})
}

// HiveRoutes implements /healthz and /v1/* API.
type HiveRoutes struct {
	handler lib.RequestHandler
	logger  lib.Logger
	env     lib.Env
	uv      *userver.Client
	users   *services.UserService
	es      *services.EnvironmentService
	ds      *services.DeviceService
	mqtt    *mqttsvc.Service
}

// NewHiveRoutes constructs HiveRoutes.
func NewHiveRoutes(
	handler lib.RequestHandler,
	logger lib.Logger,
	env lib.Env,
	uv *userver.Client,
	users *services.UserService,
	es *services.EnvironmentService,
	ds *services.DeviceService,
	mqtt *mqttsvc.Service,
) HiveRoutes {
	return HiveRoutes{
		handler: handler,
		logger:  logger,
		env:     env,
		uv:      uv,
		users:   users,
		es:      es,
		ds:      ds,
		mqtt:    mqtt,
	}
}

func (h HiveRoutes) Setup() {
	h.logger.Info("Setting up Flora Hive routes")

	ginEngine := h.handler.Gin
	ginEngine.GET("/healthz", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "flora-hive"})
	})
	ginEngine.HEAD("/healthz", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Status(http.StatusOK)
	})

	v1 := h.handler.Group
	v1.Use(maxBody(512 * 1024))
	v1.Use(middleware.AttachAuthOptional(h.env, h.uv, h.users))

	auth := v1.Group("/auth")
	auth.POST("/login", h.postLogin)
	auth.POST("/register", h.postRegister)
	auth.POST("/refresh", h.postRefresh)
	auth.POST("/logout", middleware.RequireAuth(), h.postLogout)
	auth.GET("/me", middleware.RequireAuth(), h.getMe)
	auth.PATCH("/password", middleware.RequireAuth(), h.patchPassword)
	auth.PATCH("/reset-password", middleware.RequireAuth(), h.patchPassword)

	mqttG := v1.Group("/mqtt")
	mqttG.GET("/connection", middleware.RequireAuth(), h.getMqttConnection)
	mqttG.GET("/devices", middleware.RequireAuth(), h.getMqttDevices)
	mqttG.POST("/publish", middleware.RequireAuth(), h.postMqttPublish)

	v1.GET("/environments", middleware.RequireAuth(), h.listEnvironments)
	v1.POST("/environments", middleware.RequireAuth(), h.postEnvironment)
	v1.GET("/environments/:environmentId", middleware.RequireAuth(), h.getEnvironment)
	v1.PATCH("/environments/:environmentId", middleware.RequireAuth(), h.patchEnvironment)
	v1.DELETE("/environments/:environmentId", middleware.RequireAuth(), h.deleteEnvironment)

	v1.GET("/environments/:environmentId/members", middleware.RequireAuth(), h.listMembers)
	v1.POST("/environments/:environmentId/members", middleware.RequireAuth(), h.postMember)
	v1.PATCH("/environments/:environmentId/members/:userId", middleware.RequireAuth(), h.patchMember)
	v1.DELETE("/environments/:environmentId/members/:userId", middleware.RequireAuth(), h.deleteMember)

	v1.GET("/environments/:environmentId/devices", middleware.RequireAuth(), h.listDevices)
	v1.POST("/environments/:environmentId/devices", middleware.RequireAuth(), h.postDevice)
	v1.GET("/environments/:environmentId/devices/:deviceId", middleware.RequireAuth(), h.getDevice)
	v1.PATCH("/environments/:environmentId/devices/:deviceId", middleware.RequireAuth(), h.patchDevice)
	v1.DELETE("/environments/:environmentId/devices/:deviceId", middleware.RequireAuth(), h.deleteDevice)
}

func maxBody(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		}
		c.Next()
		if c.Writer.Status() == http.StatusRequestEntityTooLarge {
			return
		}
	}
}

func pathParam(c *gin.Context, name string) string {
	return strings.TrimSpace(c.Param(name))
}

func (h HiveRoutes) postLogin(c *gin.Context) {
	if !h.env.UserverConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth_unavailable", "message": "uServer-Auth not configured"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password required"})
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password required"})
		return
	}
	out, status, msg, err := h.uv.Login(body.Username, body.Password)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "login_failed", "message": err.Error()})
		return
	}
	if out == nil {
		c.JSON(status, gin.H{"error": "login_failed", "message": msg})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h HiveRoutes) postRegister(c *gin.Context) {
	if !h.env.UserverConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth_unavailable", "message": "uServer-Auth not configured"})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  *bool  `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password required"})
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password required"})
		return
	}
	out, status, msg, err := h.uv.Register(body.Username, body.Password, body.IsAdmin)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "register_failed", "message": err.Error()})
		return
	}
	if out == nil {
		c.JSON(status, gin.H{"error": "register_failed", "message": msg})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h HiveRoutes) postRefresh(c *gin.Context) {
	if !h.env.UserverConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth_unavailable", "message": "uServer-Auth not configured"})
		return
	}
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "refresh_token required"})
		return
	}
	out, status, msg, err := h.uv.Refresh(body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "refresh_failed", "message": err.Error()})
		return
	}
	if out == nil {
		c.JSON(status, gin.H{"error": "refresh_failed", "message": msg})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h HiveRoutes) postLogout(c *gin.Context) {
	p, _ := authctx.Get(c)
	if p == nil || p.Kind != authctx.KindJWT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "JWT session required"})
		return
	}
	_, _ = h.uv.Logout(p.AccessToken)
	c.Status(http.StatusNoContent)
}

func (h HiveRoutes) getMe(c *gin.Context) {
	p, _ := authctx.Get(c)
	if p == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if p.Kind == authctx.KindAPIKey {
		c.JSON(http.StatusOK, gin.H{"kind": "api_key", "role": "service"})
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	me, status, msg, err := h.uv.Me(p.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "auth_invalid", "message": err.Error()})
		return
	}
	if me == nil {
		c.JSON(status, gin.H{"error": "auth_invalid", "message": msg})
		return
	}
	if _, err := h.users.UpsertFromMe(me); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	dbUser, err := h.users.FindByAuthUUID(me.UUID)
	if err != nil || dbUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Hive user sync failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"userver":  me,
		"hiveUser": hiveUserPublic(dbUser),
	})
}

func (h HiveRoutes) patchPassword(c *gin.Context) {
	p, _ := authctx.Get(c)
	if p == nil || p.Kind != authctx.KindJWT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Bearer access token required"})
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CurrentPassword == "" || body.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "current_password and new_password required"})
		return
	}
	status, msg, err := h.uv.ChangePassword(p.AccessToken, body.CurrentPassword, body.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "password_change_failed", "message": err.Error()})
		return
	}
	if status < 200 || status >= 300 {
		c.JSON(status, gin.H{"error": "password_change_failed", "message": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Password updated."})
}

func (h HiveRoutes) getMqttConnection(c *gin.Context) {
	c.JSON(http.StatusOK, h.mqtt.GetState())
}

func (h HiveRoutes) getMqttDevices(c *gin.Context) {
	p, _ := authctx.Get(c)
	raw := strings.TrimSpace(c.Query("include_offline"))
	includeOffline := raw == "1" || strings.EqualFold(raw, "true") || strings.EqualFold(raw, "yes")

	var allowed []string
	var allowAll bool
	switch p.Kind {
	case authctx.KindAPIKey:
		allowAll = true
	case authctx.KindJWT:
		rows, err := h.es.ListForUser(p.HiveUserID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		envIDs := make([]string, 0, len(rows))
		for _, r := range rows {
			envIDs = append(envIDs, r.ID)
		}
		ids, err := h.ds.ListRowIDsForEnvironments(envIDs)
		if err != nil {
			h.writeError(c, err)
			return
		}
		allowed = ids
	default:
		allowed = []string{}
	}

	var list []mqttsvc.PublicDevice
	if allowAll {
		list = h.mqtt.ListLiveDevices(includeOffline, nil)
	} else {
		list = h.mqtt.ListLiveDevices(includeOffline, allowed)
	}
	c.JSON(http.StatusOK, gin.H{"devices": list})
}

func (h HiveRoutes) postMqttPublish(c *gin.Context) {
	p, _ := authctx.Get(c)
	var body struct {
		Topic   string      `json:"topic"`
		Payload interface{} `json:"payload"`
		QoS     *int        `json:"qos"`
		Retain  *bool       `json:"retain"`
	}
	if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if strings.TrimSpace(body.Topic) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "topic is required"})
		return
	}
	if err := h.assertMqttPublishAllowed(p, body.Topic); err != nil {
		var fe *models.ForbiddenError
		if errors.As(err, &fe) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": fe.Message})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
		return
	}
	retain := body.Retain != nil && *body.Retain
	topic, qos, ret, n, err := h.mqtt.Publish(body.Topic, body.Payload, body.QoS, retain)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "not connected") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "mqtt_unavailable", "message": msg})
			return
		}
		if strings.Contains(strings.ToLower(msg), "qos") || strings.Contains(strings.ToLower(msg), "topic") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": msg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": msg})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"ok": true, "topic": topic, "qos": qos, "retain": ret, "bytes": n})
}

func (h HiveRoutes) assertMqttPublishAllowed(p *authctx.Principal, topic string) error {
	if p.Kind == authctx.KindAPIKey {
		return nil
	}
	if p.Kind != authctx.KindJWT {
		return &models.ForbiddenError{Message: "Authentication required"}
	}
	norm, err := mqtttopic.NormalizePublishTopic(topic, h.env.FloraTopicPrefix)
	if err != nil {
		return &models.ForbiddenError{Message: err.Error()}
	}
	deviceRowID := mqtttopic.ParseDeviceRowIDFromTopic(norm, h.env.FloraTopicPrefix)
	if deviceRowID == "" {
		return &models.ForbiddenError{Message: "Cannot derive device id from topic"}
	}
	dev, err := h.ds.GetRowByID(deviceRowID)
	if err != nil {
		return err
	}
	if dev == nil {
		return &models.ForbiddenError{Message: "Unknown device for this topic"}
	}
	_, err = h.es.RequireEnvAccess(dev.EnvironmentID, p.HiveUserID, true)
	return err
}

func (h HiveRoutes) listEnvironments(c *gin.Context) {
	p, _ := authctx.Get(c)
	if p.Kind == authctx.KindAPIKey {
		all, err := h.es.ListAll()
		if err != nil {
			h.writeError(c, err)
			return
		}
		out := make([]gin.H, 0, len(all))
		for _, e := range all {
			out = append(out, environmentPublic(&e, nil))
		}
		c.JSON(http.StatusOK, gin.H{"environments": out})
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.es.ListForUser(p.HiveUserID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		role := string(r.Role)
		out = append(out, environmentPublic(&r.Environment, &role))
	}
	c.JSON(http.StatusOK, gin.H{"environments": out})
}

func (h HiveRoutes) postEnvironment(c *gin.Context) {
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Create environment with a user JWT, not an API key"})
		return
	}
	var body struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "name required"})
		return
	}
	env, err := h.es.Create(body.Name, body.Description, p.HiveUserID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, environmentPublic(env, nil))
}

func (h HiveRoutes) getEnvironment(c *gin.Context) {
	id := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	row, err := h.es.GetByID(id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if p.Kind == authctx.KindAPIKey {
		c.JSON(http.StatusOK, environmentPublic(row, nil))
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if _, err := h.es.RequireEnvAccess(id, p.HiveUserID, false); err != nil {
		h.writeForbidden(c, err)
		return
	}
	m, _ := h.es.GetMembership(id, p.HiveUserID)
	var role *string
	if m != nil {
		r := string(m.Role)
		role = &r
	}
	c.JSON(http.StatusOK, environmentPublic(row, role))
}

type patchEnvironmentBody struct {
	Name        *string          `json:"name"`
	Description *json.RawMessage `json:"description"`
}

func (h HiveRoutes) patchEnvironment(c *gin.Context) {
	id := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	if _, err := h.es.RequireEnvAccess(id, p.HiveUserID, true); err != nil {
		h.writeForbidden(c, err)
		return
	}
	row, err := h.es.GetByID(id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	var body patchEnvironmentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	var namePtr *string
	if body.Name != nil {
		namePtr = body.Name
	}
	updateDesc := false
	var descPtr *string
	if body.Description != nil {
		raw := strings.TrimSpace(string(*body.Description))
		if raw == "" || raw == "null" {
			updateDesc = true
			descPtr = nil
		} else {
			var s string
			if err := json.Unmarshal(*body.Description, &s); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid description"})
				return
			}
			updateDesc = true
			descPtr = &s
		}
	}
	updated, err := h.es.Update(id, namePtr, updateDesc, descPtr)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, environmentPublic(updated, nil))
}

func (h HiveRoutes) deleteEnvironment(c *gin.Context) {
	id := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	if _, err := h.es.RequireEnvAccess(id, p.HiveUserID, true); err != nil {
		h.writeForbidden(c, err)
		return
	}
	row, err := h.es.GetByID(id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if err := h.es.Delete(id); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h HiveRoutes) listMembers(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	if p.Kind == authctx.KindAPIKey {
		row, err := h.es.GetByID(envID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		if row == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		members, err := h.es.ListMembers(envID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"members": members})
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if _, err := h.es.RequireEnvAccess(envID, p.HiveUserID, false); err != nil {
		h.writeForbidden(c, err)
		return
	}
	members, err := h.es.ListMembers(envID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h HiveRoutes) postMember(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	if _, err := h.es.RequireEnvAccess(envID, p.HiveUserID, true); err != nil {
		h.writeForbidden(c, err)
		return
	}
	var body struct {
		AuthUserUUID string `json:"authUserUuid"`
		Role         string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if strings.TrimSpace(body.AuthUserUUID) == "" || (body.Role != "viewer" && body.Role != "editor") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "authUserUuid and role (viewer|editor) required"})
		return
	}
	target, err := h.users.FindByAuthUUID(body.AuthUserUUID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Hive user not found for that auth UUID; the user must call GET /v1/auth/me once after registering.",
		})
		return
	}
	role := models.MemberRole(body.Role)
	if err := h.es.UpsertMember(envID, target.ID, role); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"environmentId": envID, "userId": target.ID, "role": body.Role})
}

func (h HiveRoutes) patchMember(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	userID := pathParam(c, "userId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	if _, err := h.es.RequireEnvAccess(envID, p.HiveUserID, true); err != nil {
		h.writeForbidden(c, err)
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if body.Role != "viewer" && body.Role != "editor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "role required"})
		return
	}
	if err := h.es.UpsertMember(envID, userID, models.MemberRole(body.Role)); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"environmentId": envID, "userId": userID, "role": body.Role})
}

func (h HiveRoutes) deleteMember(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	userID := pathParam(c, "userId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	if _, err := h.es.RequireEnvAccess(envID, p.HiveUserID, true); err != nil {
		h.writeForbidden(c, err)
		return
	}
	if err := h.es.RemoveMember(envID, userID); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parentFilterFromQuery(c *gin.Context) *ports.ParentDeviceFilter {
	raw, ok := c.GetQuery("parent")
	if !ok {
		return ports.NewParentFilterAll()
	}
	if raw == "null" || raw == "" {
		return ports.NewParentFilterRoot()
	}
	return ports.NewParentFilterID(strings.TrimSpace(raw))
}

func (h HiveRoutes) listDevices(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	pf := parentFilterFromQuery(c)
	if p.Kind == authctx.KindAPIKey {
		env, err := h.es.GetByID(envID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		if env == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		rows, err := h.ds.ListByEnvironment(envID, pf)
		if err != nil {
			h.writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"devices": mapDevices(rows)})
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.ds.ListInEnvironment(envID, p.HiveUserID, pf)
	if err != nil {
		h.writeForbiddenOrError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"devices": mapDevices(rows)})
}

func (h HiveRoutes) postDevice(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	var body struct {
		DeviceType     string  `json:"deviceType"`
		DeviceID       string  `json:"deviceId"`
		DisplayName    *string `json:"displayName"`
		ParentDeviceID *string `json:"parentDeviceId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if strings.TrimSpace(body.DeviceType) == "" || strings.TrimSpace(body.DeviceID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "deviceType and deviceId required"})
		return
	}
	parent := normalizeParentPtr(body.ParentDeviceID)
	row, err := h.ds.CreateDevice(envID, p.HiveUserID, body.DeviceType, body.DeviceID, body.DisplayName, parent)
	if err != nil {
		if errors.Is(err, services.ErrInvalidParent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
			return
		}
		h.writeForbiddenOrError(c, err)
		return
	}
	c.JSON(http.StatusCreated, devicePublic(row))
}

func normalizeParentPtr(p *string) *string {
	if p == nil {
		return nil
	}
	if strings.TrimSpace(*p) == "" {
		return nil
	}
	return p
}

func (h HiveRoutes) getDevice(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	logicalID := pathParam(c, "deviceId")
	p, _ := authctx.Get(c)
	if p.Kind == authctx.KindAPIKey {
		row, err := h.ds.GetRowByEnvAndDeviceID(envID, logicalID)
		if err != nil {
			h.writeError(c, err)
			return
		}
		if row == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusOK, devicePublic(row))
		return
	}
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	row, err := h.ds.GetByEnvAndDeviceID(envID, logicalID, p.HiveUserID)
	if err != nil {
		h.writeForbiddenOrError(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, devicePublic(row))
}

type patchDeviceBody struct {
	DeviceType     *string          `json:"deviceType"`
	DeviceID       *string          `json:"deviceId"`
	DisplayName    *json.RawMessage `json:"displayName"`
	ParentDeviceID *json.RawMessage `json:"parentDeviceId"`
}

func (h HiveRoutes) patchDevice(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	logicalID := pathParam(c, "deviceId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	var body patchDeviceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	updateDisplay := false
	var displayPtr *string
	if body.DisplayName != nil {
		raw := strings.TrimSpace(string(*body.DisplayName))
		if raw == "" || raw == "null" {
			updateDisplay = true
			displayPtr = nil
		} else {
			var s string
			if err := json.Unmarshal(*body.DisplayName, &s); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid displayName"})
				return
			}
			updateDisplay = true
			displayPtr = &s
		}
	}
	updateParent := false
	var parentPtr *string
	if body.ParentDeviceID != nil {
		raw := strings.TrimSpace(string(*body.ParentDeviceID))
		if raw == "" || raw == "null" {
			updateParent = true
			parentPtr = nil
		} else {
			var s string
			if err := json.Unmarshal(*body.ParentDeviceID, &s); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid parentDeviceId"})
				return
			}
			updateParent = true
			parentPtr = &s
		}
	}
	row, err := h.ds.UpdateByEnvAndDeviceID(envID, logicalID, p.HiveUserID, body.DeviceType, body.DeviceID, updateDisplay, displayPtr, updateParent, parentPtr)
	if err != nil {
		h.writeForbiddenOrError(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, devicePublic(row))
}

func (h HiveRoutes) deleteDevice(c *gin.Context) {
	envID := pathParam(c, "environmentId")
	logicalID := pathParam(c, "deviceId")
	p, _ := authctx.Get(c)
	if p.Kind != authctx.KindJWT {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "JWT required"})
		return
	}
	ok, err := h.ds.DeleteByEnvAndDeviceID(envID, logicalID, p.HiveUserID)
	if err != nil {
		h.writeForbiddenOrError(c, err)
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h HiveRoutes) writeError(c *gin.Context, err error) {
	h.logger.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
}

func (h HiveRoutes) writeForbidden(c *gin.Context, err error) {
	var fe *models.ForbiddenError
	if errors.As(err, &fe) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": fe.Message})
		return
	}
	h.writeError(c, err)
}

func (h HiveRoutes) writeForbiddenOrError(c *gin.Context, err error) {
	var fe *models.ForbiddenError
	if errors.As(err, &fe) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": fe.Message})
		return
	}
	h.writeError(c, err)
}

func environmentPublic(e *models.Environment, role *string) gin.H {
	h := gin.H{
		"id":          e.ID,
		"name":        e.Name,
		"path":        "environments/" + e.ID,
		"description": e.Description,
		"createdAt":   e.CreatedAt,
		"updatedAt":   e.UpdatedAt,
	}
	if role != nil {
		h["role"] = *role
	}
	return h
}

func devicePublic(d *models.Device) gin.H {
	return gin.H{
		"id":             d.ID,
		"path":           "environments/" + d.EnvironmentID + "/devices/" + d.DeviceID,
		"environmentId":  d.EnvironmentID,
		"parentDeviceId": d.ParentDeviceID,
		"deviceType":     d.DeviceType,
		"deviceId":       d.DeviceID,
		"displayName":    d.DisplayName,
		"createdAt":      d.CreatedAt,
		"updatedAt":      d.UpdatedAt,
	}
}

func mapDevices(rows []models.Device) []gin.H {
	out := make([]gin.H, 0, len(rows))
	for i := range rows {
		out = append(out, devicePublic(&rows[i]))
	}
	return out
}

func hiveUserPublic(u *models.HiveUser) gin.H {
	return gin.H{
		"id":         u.ID,
		"authUuid":   u.AuthUUID,
		"username":   u.Username,
		"systemName": u.SystemName,
		"updatedAt":  u.UpdatedAt,
	}
}
