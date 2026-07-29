package web

import (
	"github.com/gin-gonic/gin"

	"sipon-api/internal/app/apperror"
	santriUsecase "sipon-api/internal/app/usecase/santri"
	"sipon-api/internal/interfaces/http/httperror"
	"sipon-api/internal/interfaces/http/middleware"
	"sipon-api/internal/interfaces/http/respond"
)

type SantriHandler struct {
	uc *santriUsecase.UseCases
}

func NewSantriHandler(uc *santriUsecase.UseCases) *SantriHandler {
	return &SantriHandler{uc: uc}
}

func userIDFromCtx(c *gin.Context) string {
	p := middleware.GetPrincipal(c)
	if p == nil {
		return ""
	}
	return p.UserID
}

// ── Santri Profile ────────────────────────────────────────────────────────────

func (h *SantriHandler) GetSantri(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	resp, err := h.uc.GetSantri.Execute(c.Request.Context(), uid)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "data santri ditemukan", resp)
}

func (h *SantriHandler) UpdateSantri(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	var req santriUsecase.UpdateSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.uc.UpdateSantri.Execute(c.Request.Context(), uid, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *SantriHandler) RequestSantri(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	resp, err := h.uc.RequestSantri.Execute(c.Request.Context(), uid)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, resp.Message, resp)
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (h *SantriHandler) CreateSantri(c *gin.Context) {
	var req santriUsecase.CreateSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.uc.CreateSantri.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "santri berhasil dibuat", resp)
}

func (h *SantriHandler) ListSantri(c *gin.Context) {
	var query santriUsecase.ListSantriQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httperror.Handle(c, err)
		return
	}
	data, meta, err := h.uc.ListSantri.Execute(c.Request.Context(), query)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "data santri ditemukan", data, meta)
}

func (h *SantriHandler) ListSantriRequests(c *gin.Context) {
	var query santriUsecase.ListSantriRequestsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httperror.Handle(c, err)
		return
	}
	data, meta, err := h.uc.ListSantriRequests.Execute(c.Request.Context(), query)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "data request ditemukan", data, meta)
}

func (h *SantriHandler) ApproveSantriRequest(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	requestID := c.Param("id")
	if requestID == "" {
		httperror.Handle(c, apperror.BadRequest("request id required"))
		return
	}
	var req santriUsecase.ApproveSantriRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.uc.ApproveSantriRequest.Execute(c.Request.Context(), requestID, uid, req); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "request disetujui", nil)
}

func (h *SantriHandler) RejectSantriRequest(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	requestID := c.Param("id")
	if requestID == "" {
		httperror.Handle(c, apperror.BadRequest("request id required"))
		return
	}
	var req santriUsecase.RejectSantriRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.uc.RejectSantriRequest.Execute(c.Request.Context(), requestID, uid, req); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "request ditolak", nil)
}

// ── Dokumen ───────────────────────────────────────────────────────────────────

func (h *SantriHandler) DokumenPresign(c *gin.Context) {
	var req santriUsecase.DokumenPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.uc.DokumenPresign.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "presigned URL dibuat", resp)
}

func (h *SantriHandler) DokumenConfirm(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	var req santriUsecase.DokumenConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.uc.DokumenConfirm.Execute(c.Request.Context(), uid, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "dokumen berhasil disimpan", resp)
}

func (h *SantriHandler) DokumenList(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	kind := c.Query("kind")
	resp, err := h.uc.DokumenList.Execute(c.Request.Context(), uid, kind)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen ditemukan", resp)
}

func (h *SantriHandler) DokumenAccess(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	dokumenID := c.Param("id")
	if dokumenID == "" {
		httperror.Handle(c, apperror.BadRequest("dokumen_id required"))
		return
	}
	resp, err := h.uc.DokumenAccess.Execute(c.Request.Context(), uid, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "access URL dibuat", resp)
}

func (h *SantriHandler) DokumenDelete(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	dokumenID := c.Param("id")
	if dokumenID == "" {
		httperror.Handle(c, apperror.BadRequest("dokumen_id required"))
		return
	}
	if err := h.uc.DokumenDelete.Execute(c.Request.Context(), uid, dokumenID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen berhasil dihapus", nil)
}

func (h *SantriHandler) DokumenVerify(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	dokumenID := c.Param("id")
	if dokumenID == "" {
		httperror.Handle(c, apperror.BadRequest("dokumen_id required"))
		return
	}
	if err := h.uc.DokumenVerify.Execute(c.Request.Context(), dokumenID, uid); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen berhasil diverifikasi", nil)
}

func (h *SantriHandler) DokumenReject(c *gin.Context) {
	uid := userIDFromCtx(c)
	if uid == "" {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	dokumenID := c.Param("id")
	if dokumenID == "" {
		httperror.Handle(c, apperror.BadRequest("dokumen_id required"))
		return
	}
	var req santriUsecase.VerifyDokumenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.uc.DokumenReject.Execute(c.Request.Context(), dokumenID, uid, req.Notes); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen ditolak", nil)
}
